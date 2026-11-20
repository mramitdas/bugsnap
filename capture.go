package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var depNames = map[string]bool{
	"go.mod": true, "go.sum": true,
	"package.json": true, "package-lock.json": true, "yarn.lock": true, "pnpm-lock.yaml": true, "bun.lockb": true,
	"requirements.txt": true, "pyproject.toml": true, "poetry.lock": true, "Pipfile.lock": true,
	"Cargo.toml": true, "Cargo.lock": true,
	"Gemfile": true, "Gemfile.lock": true,
	"composer.json": true, "composer.lock": true,
	"pom.xml": true, "build.gradle": true, "build.gradle.kts": true,
	"mix.exs": true, "mix.lock": true,
}

var skipDirs = map[string]bool{
	"node_modules": true, ".git": true, "vendor": true, "dist": true,
	"build": true, "target": true, ".venv": true, "venv": true,
	"bin": true, "obj": true, "Pods": true, ".next": true, "coverage": true,
}

const (
	maxFile     = 2 << 20
	maxUntr     = 5 << 20
	maxLog      = 256 << 10
	maxLines    = 200
	maxDepDepth = 3
	maxDepTotal = 100 << 20
)

func cmdCapture(args []string) error {
	fs := flag.NewFlagSet("capture", flag.ExitOnError)
	logPath := fs.String("log", "", "log file to include (last --tail lines)")
	tail := fs.Int("tail", maxLines, "log lines to keep")
	outDir := fs.String("out", ".", "output directory")
	note := fs.String("note", "", "note to attach to the bundle")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cmd := fs.Args()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	files := map[string][]byte{}
	m := Manifest{
		Version:   1,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Host:      hostInfo(),
		Note:      *note,
	}
	m.Command.Cwd = cwd

	gitFiles, gi := captureGit(cwd)
	m.Git = gi
	for k, v := range gitFiles {
		files[k] = v
	}
	for k, v := range captureDeps(cwd) {
		files[k] = v
	}

	files["env/environment.txt"] = []byte(strings.Join(redactEnv(os.Environ()), "\n") + "\n")

	if *logPath != "" {
		b, err := tailFile(*logPath, *tail)
		if err != nil {
			return err
		}
		files["logs/"+filepath.Base(*logPath)] = b
	}

	if len(cmd) > 0 {
		m.Command.Argv = cmd
		start := time.Now()
		var buf bytes.Buffer
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = cwd
		c.Stdout = &buf
		c.Stderr = &buf
		runErr := c.Run()
		dur := time.Since(start)

		exit := 0
		if runErr != nil {
			exit = -1
			if ee, ok := runErr.(*exec.ExitError); ok {
				exit = ee.ExitCode()
			}
		}
		m.Command.ExitCode = &exit
		m.Command.DurationMS = dur.Milliseconds()

		hdr := fmt.Sprintf("$ %s\n# exit=%d duration=%s\n\n", strings.Join(cmd, " "), exit, dur.Round(time.Millisecond))
		files["logs/command.log"] = append([]byte(hdr), tailBytes(&buf, maxLog)...)
	}

	m.Files = sortedFiles(files)
	mb, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	files["manifest.json"] = mb

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}
	name := fmt.Sprintf("bugsnap-%s.zip", time.Now().UTC().Format("20060102T150405Z"))
	path := filepath.Join(*outDir, name)
	if err := writeBundle(path, files); err != nil {
		return err
	}

	fmt.Printf("captured %s\n", path)
	fmt.Printf("  host:    %s/%s\n", m.Host.OS, m.Host.Arch)
	fmt.Printf("  git:     %s\n", shortSHA(m.Git.SHA))
	fmt.Printf("  files:   %d\n", len(m.Files))
	if len(cmd) > 0 {
		fmt.Printf("  command: exit %d\n", *m.Command.ExitCode)
	}
	return nil
}

func hostInfo() HostInfo {
	h, _ := os.Hostname()
	return HostInfo{OS: runtime.GOOS, Arch: runtime.GOARCH, Hostname: h}
}

func gitOut(dir string, args ...string) string {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func captureGit(dir string) (map[string][]byte, GitInfo) {
	files := map[string][]byte{}
	var gi GitInfo
	if strings.TrimSpace(gitOut(dir, "rev-parse", "--is-inside-work-tree")) != "true" {
		return files, gi
	}

	gi.SHA = strings.TrimSpace(gitOut(dir, "rev-parse", "HEAD"))
	gi.Branch = strings.TrimSpace(gitOut(dir, "branch", "--show-current"))
	gi.Remote = strings.TrimSpace(gitOut(dir, "remote", "get-url", "origin"))

	status := gitOut(dir, "status", "--porcelain")
	gi.Dirty = strings.TrimSpace(status) != ""
	files["git/status.txt"] = []byte(status)

	diff := gitOut(dir, "diff", "HEAD")
	if diff == "" {
		diff = gitOut(dir, "diff") + gitOut(dir, "diff", "--cached")
	}
	files["git/diff.patch"] = []byte(diff)

	untracked := strings.TrimSpace(gitOut(dir, "ls-files", "--others", "--exclude-standard"))
	if untracked != "" {
		files["git/untracked.txt"] = []byte(untracked + "\n")
		var total int64
		for _, p := range strings.Split(untracked, "\n") {
			if total >= maxUntr {
				break
			}
			if isSecretFile(p) {
				continue
			}
			full := filepath.Join(dir, p)
			info, err := os.Stat(full)
			if err != nil || info.IsDir() || info.Size() > maxFile {
				continue
			}
			b, err := os.ReadFile(full)
			if err != nil {
				continue
			}
			files["git/untracked/"+filepath.ToSlash(p)] = b
			total += int64(len(b))
		}
	}
	return files, gi
}

func captureDeps(root string) map[string][]byte {
	out := map[string][]byte{}
	var total int64
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		slash := filepath.ToSlash(rel)
		depth := 0
		if slash != "." {
			depth = strings.Count(slash, "/") + 1
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			if skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".") || depth > maxDepDepth {
				return fs.SkipDir
			}
			return nil
		}
		if depth > maxDepDepth || !depNames[d.Name()] || total >= maxDepTotal {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > maxFile {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		out["deps/"+slash] = b
		total += int64(len(b))
		return nil
	})
	return out
}

func tailFile(path string, n int) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	if n > 0 && len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return []byte(strings.Join(lines, "\n")), nil
}

func tailBytes(buf *bytes.Buffer, n int) []byte {
	data := buf.Bytes()
	if len(data) <= n {
		return data
	}
	data = data[len(data)-n:]
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		data = data[i+1:]
	}
	return data
}

func shortSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	if s == "" {
		return "(none)"
	}
	return s
}
