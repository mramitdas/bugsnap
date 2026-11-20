# bugsnap

**Stop chasing "it works on my machine". Capture a bug once, replay it anywhere.**

[![ci](https://github.com/mramitdas/bugsnap/actions/workflows/ci.yml/badge.svg)](https://github.com/mramitdas/bugsnap/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/mramitdas/bugsnap)](https://github.com/mramitdas/bugsnap/releases)
[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Reproducing a bug locally is still the slowest part of debugging. A 2026 survey of Java developers found **60% call reproducing bugs locally their biggest time sink**. The context lives in git, env vars, logs, dependency files, and the exact command that failed. `bugsnap` packs all of it into one bundle you can hand to a teammate — or replay in Docker.

```
bugsnap capture -- go test ./...
# captured bugsnap-20261120T183500Z.zip

bugsnap inspect bugsnap-20261120T183500Z.zip
bugsnap replay  bugsnap-20261120T183500Z.zip --apply --run
bugsnap replay  bugsnap-20261120T183500Z.zip --apply --docker
```

## Install

```sh
go install github.com/mramitdas/bugsnap@latest
```

Or download a binary from [Releases](https://github.com/mramitdas/bugsnap/releases). Or build from source:

```sh
git clone https://github.com/mramitdas/bugsnap
cd bugsnap && go build -o bugsnap .
```

`bugsnap` is a single static binary with **zero dependencies**.

## What gets captured

| Part | Contents |
|------|----------|
| `git/` | commit SHA, branch, remote, porcelain status, working-tree diff, untracked files |
| `env/` | environment variables, **secrets redacted** |
| `deps/` | lockfiles for Go, Node, Python, Rust, Ruby, PHP, Java, and Elixir — including nested monorepo packages |
| `logs/` | your `--log` file (tail) and the captured command's output |
| `manifest.json` | everything indexed: host, git state, command, exit code, duration |

## Commands

```
bugsnap capture [flags] [-- COMMAND [ARGS...]]
bugsnap inspect BUNDLE.zip
bugsnap replay  BUNDLE.zip [--apply] [--run] [--docker] [--image IMAGE] [--dir DIR]
```

### capture

```sh
# capture state and run a command
bugsnap capture -- npm test

# include the tail of an application log
bugsnap capture --log app.log --tail 500 -- ./run.sh

# attach context for the person who will replay it
bugsnap capture --note "checkout fails on second run" -- go test ./...
```

### inspect

Read-only summary: git state, command, exit code, files, and an environment diff between the bundle and your machine.

### replay

- default: dry run (summary + env diff, changes nothing)
- `--apply`: restore the git patch, untracked files, and lockfiles into `--dir`
- `--run`: re-execute the captured command in `--dir`; the original exit code is preserved
- `--docker`: run the command inside a container instead of your host
  - builds `Dockerfile` from `--dir`, or uses `--image IMAGE`
  - mounts the target directory at `/work`

Be deliberate: `--apply`, `--run`, and `--docker` change files and execute code. Inspect first.

## Security

- Secret-looking keys (`*TOKEN*`, `*KEY*`, `*SECRET*`, `*PASSWORD*`, ...) are masked as `[REDACTED]`.
- Token-shaped values (`ghp_`, `sk-`, `AKIA`, `xoxb-`, ...) are masked too.
- Well-known credential files (`.env*`, `*.pem`, `*.key`, `id_rsa*`, `.npmrc`, ...) are never included as untracked files.
- Bundles are plain zip files. **Review before sharing.** No upload, no telemetry.
- Replay refuses path traversal outside the target directory.

## Limitations

- Dependency discovery walks up to 3 directory levels, skipping `node_modules`, `vendor`, `.venv`, and build output.
- Docker replay mounts your working tree; it does not reproduce the original image or OS.
- Untracked files over 1 MB and totals over 5 MB are skipped.

## Why this exists

Most "repro" tools assume a hosted error tracker. `bugsnap` is local-first and works with any language, any CI, and any teammate. If it saves you one round of "can you send me your logs and your commit?", it earned its keep.

## License

MIT. Not affiliated with SmartBear BugSnag.
