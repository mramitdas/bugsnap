# reprokit

**Stop chasing "it works on my machine". Capture a bug once, replay it anywhere.**

[![ci](https://github.com/mramitdas/reprokit/actions/workflows/ci.yml/badge.svg)](https://github.com/mramitdas/reprokit/actions/workflows/ci.yml)
[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Reproducing a bug locally is still the slowest part of debugging. A 2026 survey of Java developers found **60% call reproducing bugs locally their biggest time sink**. Context lives in git, in env vars, in logs, in dependency files, and in the exact command that failed. `reprokit` packs all of it into one bundle you can hand to a teammate.

```
reprokit capture -- go test ./...
# captured reprokit-20261120T183500Z.zip

reprokit inspect reprokit-20261120T183500Z.zip
reprokit replay  reprokit-20261120T183500Z.zip --apply --run
```

## Install

```sh
go install github.com/mramitdas/reprokit@latest
```

Or build from source:

```sh
git clone https://github.com/mramitdas/reprokit
cd reprokit && go build -o reprokit .
```

`reprokit` is a single static binary with **zero dependencies**.

## What gets captured

| Part | Contents |
|------|----------|
| `git/` | commit SHA, branch, remote, porcelain status, working-tree diff, untracked files |
| `env/` | environment variables, **secrets redacted** |
| `deps/` | lockfiles for Go, Node, Python, Rust, Ruby, PHP, Java, Elixir |
| `logs/` | your `--log` file (tail) and the captured command's output |
| `manifest.json` | everything indexed: host, git state, command, exit code, duration |

## Commands

```
reprokit capture [flags] [-- COMMAND [ARGS...]]
reprokit inspect BUNDLE.zip
reprokit replay  BUNDLE.zip [--apply] [--run] [--dir DIR]
```

### capture

```sh
# capture state and run a command
reprokit capture -- npm test

# include the tail of an application log
reprokit capture --log app.log --tail 500 -- ./run.sh

# attach context for the person who will replay it
reprokit capture --note "checkout fails on second run" -- go test ./...
```

### inspect

Read-only summary: git state, command, exit code, files, and an environment diff between the bundle and your machine.

### replay

- default: dry run (summary + env diff, changes nothing)
- `--apply`: restore the git patch, untracked files, and lockfiles into `--dir`
- `--run`: re-execute the captured command in `--dir`

Be deliberate: `--apply` and `--run` change files and execute code. Inspect first.

## Security

- Secret-looking keys (`*TOKEN*`, `*KEY*`, `*SECRET*`, `*PASSWORD*`, ...) are masked as `[REDACTED]`.
- Token-shaped values (`ghp_`, `sk-`, `AKIA`, `xoxb-`, ...) are masked too.
- Well-known credential files (`.env*`, `*.pem`, `*.key`, `id_rsa*`, `.npmrc`, ...) are never included as untracked files.
- Bundles are plain zip files. **Review before sharing.** No upload, no telemetry.
- Replay refuses path traversal outside the target directory.

## Limitations (v0)

- Root-level dependency files only; monorepo workspace files are not walked.
- No Docker/container replay; `--run` uses your host environment.
- Untracked files over 1 MB and totals over 5 MB are skipped.

## Why this exists

Most "repro" tools assume a hosted error tracker. `reprokit` is local-first and works with any language, any CI, and any teammate. If it saves you one round of "can you send me your logs and your commit?", it earned its keep.

## License

MIT
