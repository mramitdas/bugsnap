package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func cmdInspect(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: reprokit inspect BUNDLE.zip")
	}
	m, files, err := readBundle(args[0])
	if err != nil {
		return err
	}
	printSummary(m)
	printEnvDiff(files)

	fmt.Println("\nFiles:")
	for _, name := range m.Files {
		fmt.Printf("  %-32s %d bytes\n", name, len(files[name]))
	}
	return nil
}

func cmdReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	apply := fs.Bool("apply", false, "apply git patch and restore dependency files")
	run := fs.Bool("run", false, "re-run the captured command")
	dir := fs.String("dir", ".", "target directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("usage: reprokit replay BUNDLE.zip [--apply] [--run] [--dir DIR]")
	}
	bundle := rest[0]
	if len(rest) > 1 {
		// allow flags after the bundle path
		if err := fs.Parse(rest[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("usage: reprokit replay BUNDLE.zip [--apply] [--run] [--dir DIR]")
		}
	}

	m, files, err := readBundle(bundle)
	if err != nil {
		return err
	}
	printSummary(m)
	printEnvDiff(files)

	if !*apply && !*run {
		fmt.Println("\ndry run: pass --apply to restore files, --run to execute the captured command")
		return nil
	}

	if *apply {
		if err := applyBundle(files, *dir); err != nil {
			return err
		}
		fmt.Println("\napplied git patch and restored dependency files")
	}

	if *run {
		if len(m.Command.Argv) == 0 {
			return fmt.Errorf("bundle has no captured command")
		}
		fmt.Printf("\nrunning: %s\n", strings.Join(m.Command.Argv, " "))
		c := exec.Command(m.Command.Argv[0], m.Command.Argv[1:]...)
		c.Dir = *dir
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("replay command failed: %w", err)
		}
	}
	return nil
}

func printSummary(m Manifest) {
	fmt.Printf("reprokit bundle (v%d)\n", m.Version)
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
		tmp, err := os.CreateTemp("", "reprokit-*.patch")
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
