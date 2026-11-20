# Contributing to bugsnap

Thanks for helping make bug reproduction less painful.

## Development

```sh
git clone https://github.com/mramitdas/bugsnap
cd bugsnap
go build -o bugsnap .
./bugsnap version
```

Requirements: Go 1.22+.

## Before opening a PR

```sh
gofmt -l .      # should print nothing
go vet ./...
go test ./...
```

- Keep the zero-dependency policy: standard library only unless there is a strong reason.
- Add a test for new logic (parsing, redaction, bundle handling).
- Keep changes small and focused; one feature or fix per PR.
- Do not weaken secret redaction or path-traversal guards.
- Use clear commit messages; `feat:`, `fix:`, `docs:`, `test:`, `chore:` are welcome.

## Reporting bugs

Include the output of `bugsnap version`, your OS/arch, and the command you ran. Do **not** paste real secrets or full environment dumps — share a repro bundle with `bugsnap capture` instead.

## Security

See [SECURITY.md](SECURITY.md). Report vulnerabilities privately.
