# Security

## Reporting a vulnerability

Please report security issues privately through GitHub's
[private vulnerability reporting](https://github.com/mramitdas/bugsnap/security/advisories/new)
rather than a public issue. We aim to respond within 72 hours.

## What bugsnap handles

- Bundles contain git metadata, redacted environment variables, dependency files, logs, and untracked source files.
- Secret-looking env keys and token-shaped values are masked, and well-known credential files are excluded.
- No data is uploaded anywhere and bugsnap sends no telemetry.
- Replay refuses to write outside the target directory.

Bundles are ordinary zip archives. Review them before sharing. If you find a
redaction bypass, treat it as a security issue and report it privately.
