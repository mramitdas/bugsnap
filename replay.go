package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type replayResult struct {
	Bundle   string   `json:"bundle"`
	Applied  bool     `json:"applied"`
	Ran      bool     `json:"ran"`
	ExitCode *int     `json:"exit_code,omitempty"`
	Manifest Manifest `json:"manifest"`
}

func cmdInspect(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "print machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("usage: bugsnap inspect BUNDLE.zip [--json]")
	}
	bundle := rest[0]
	if len(rest) > 1 {
		if err := fs.Parse(rest[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("usage: bugsnap inspect BUNDLE.zip [--json]")
		}
	}

	m, files, err := readBundle(bundle)
	if err != nil {
		return err
	}
	if *jsonOut {
		sizes := make(map[string]int, len(m.Files))
		for _, name := range m.Files {
			sizes[name] = len(files[name])
		}
		return emitJSON(struct {
			Bundle   string         `json:"bundle"`
			Manifest Manifest       `json:"manifest"`
			Files    map[string]int `json:"files"`
		}{bundle, m, sizes})
	}

	printSummary(m)
	printEnvDiff(files)
	fmt.Println("\nFiles:")
	for _, name := range m.Files {
		fmt.Printf("  %-40s %d bytes\n", name, len(files[name]))
	}
	return nil
}

func cmdReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	apply := fs.Bool("apply", false, "apply git patch and restore dependency files")
	run := fs.Bool("run", false, "re-run the captured command")
	docker := fs.Bool("docker", false, "re-run inside Docker (builds Dockerfile in --dir, or use --image)")
	image := fs.String("image", "", "docker image for --docker instead of building a Dockerfile")
	dir := fs.String("dir", ".", "target directory")
	jsonOut := fs.Bool("json", false, "print machine-readable JSON (child output goes to stderr)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("usage: bugsnap replay BUNDLE.zip [--apply] [--run] [--docker] [--json] [--dir DIR]")
	}
	bundle := rest[0]
	if len(rest) > 1 {
		// allow flags after the bundle path
		if err := fs.Parse(rest[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("usage: bugsnap replay BUNDLE.zip [--apply] [--run] [--docker] [--json] [--dir DIR]")
		}
	}

	m, files, err := readBundle(bundle)
	if err != nil {
		return err
	}
	if !*jsonOut {
		printSummary(m)
		printEnvDiff(files)
	}

	if !*apply && !*run && !*docker {
		if *jsonOut {
			return emitJSON(replayResult{Bundle: bundle, Manifest: m})
		}
		fmt.Println("\ndry run: pass --apply to restore files, --run to execute, --docker to run in a container")
		return nil
	}

	if *apply {
		if err := applyBundle(files, *dir); err != nil {
			return err
		}
		if !*jsonOut {
			fmt.Println("\napplied git patch and restored dependency files")
		}
	}

	if len(m.Command.Argv) == 0 {
		return fmt.Errorf("bundle has no captured command")
	}

	ran := *run || *docker
	var runErr error
	if *docker {
		runErr = runDocker(m.Command.Argv, *dir, *image, *jsonOut)
	} else if *run {
		if !*jsonOut {
			fmt.Printf("\nrunning: %s\n", strings.Join(m.Command.Argv, " "))
		}
		runErr = runLocal(m.Command.Argv, *dir, *jsonOut)
	}

	code := 0
	if runErr != nil {
		var ee exitError
		if !errors.As(runErr, &ee) {
			return runErr
		}
		code = ee.code
	}

	if *jsonOut {
		res := replayResult{Bundle: bundle, Applied: *apply, Ran: ran, Manifest: m}
		if ran {
			res.ExitCode = &code
		}
		if err := emitJSON(res); err != nil {
			return err
		}
	}
	if runErr != nil {
		return exitError{code: code}
	}
	return nil
}

func runLocal(argv []string, dir string, jsonMode bool) error {
	c := exec.Command(argv[0], argv[1:]...)
	c.Dir = dir
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	if jsonMode {
		c.Stdout = os.Stderr
	}
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return exitError{code: exitCodeOf(err)}
	}
	return nil
}

func runDocker(argv []string, dir, image string, jsonMode bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	logw := os.Stdout
	if jsonMode {
		logw = os.Stderr
	}

	if image == "" {
		if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err != nil {
			return fmt.Errorf("--docker needs a Dockerfile in %s (or pass --image)", abs)
		}
		image = "bugsnap-replay"
		fmt.Fprintf(logw, "\nbuilding docker image from %s/Dockerfile\n", abs)
		build := exec.Command("docker", "build", "-t", image, abs)
		build.Stdout = logw
		build.Stderr = os.Stderr
		if err := build.Run(); err != nil {
			return fmt.Errorf("docker build: %w", err)
		}
	}

	fmt.Fprintf(logw, "\nrunning in docker (%s): %s\n", image, strings.Join(argv, " "))
	args := []string{"run", "--rm", "-v", abs + ":/work", "-w", "/work", image}
	args = append(args, argv...)
	c := exec.Command("docker", args...)
	c.Stdin = os.Stdin
	c.Stdout = logw
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return exitError{code: exitCodeOf(err)}
	}
	return nil
}

func exitCodeOf(err error) int {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return 1
}

func printSummary(m Manifest) {
	fmt.Printf("bugsnap bundle (v%d)\n", m.Version)
	fmt.Printf("  created: %s\n", m.CreatedAt)
	fmt.Printf("  host:    %s/%s (%s)\n", m.Host.OS, m.Host.Arch, m.Host.Hostname)
	if m.Git.SHA != "" {
		dirty := "clean"
		if m.Git.Dirty {
			dirty = "dirty"
		}
		fmt.Printf("  git:     %s (%s, %s)\n", shortSHA(m.Git.SHA), m.Git.Branch, dirty)
	}
	if len(m.Command.Argv) > 0 {
		ec := -1
		if m.Command.ExitCode != nil {
			ec = *m.Command.ExitCode
		}
		fmt.Printf("  command: %s  exit=%d  %dms\n", strings.Join(m.Command.Argv, " "), ec, m.Command.DurationMS)
	}
	if m.Note != "" {
		fmt.Printf("  note:    %s\n", m.Note)
	}
}

func printEnvDiff(files map[string][]byte) {
	b, ok := files["env/environment.txt"]
	if !ok {
		return
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	fmt.Println("\nEnvironment diff (bundle -> current):")
	fmt.Print(diffEnv(parseEnv(lines), parseEnv(redactEnv(os.Environ()))))
}

func applyBundle(files map[string][]byte, dir string) error {
	if patch := files["git/diff.patch"]; len(patch) > 0 {
		tmp, err := os.CreateTemp("", "bugsnap-*.patch")
		if err != nil {
			return err
		}
		defer os.Remove(tmp.Name())
		if _, err := tmp.Write(patch); err != nil {
			tmp.Close()
			return err
		}
		tmp.Close()

		c := exec.Command("git", "apply", "--whitespace=nowarn", tmp.Name())
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			return fmt.Errorf("git apply: %v: %s", err, strings.TrimSpace(string(out)))
		}
	}

	for name, b := range files {
		if rel, ok := strings.CutPrefix(name, "git/untracked/"); ok {
			if err := safeWrite(dir, rel, b); err != nil {
				return err
			}
		}
	}
	for name, b := range files {
		if rel, ok := strings.CutPrefix(name, "deps/"); ok {
			if err := safeWrite(dir, rel, b); err != nil {
				return err
			}
		}
	}
	return nil
}

func safeWrite(base, rel string, data []byte) error {
	clean := filepath.Clean(rel)
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("refusing unsafe path %q", rel)
	}
	full := filepath.Join(base, clean)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}
