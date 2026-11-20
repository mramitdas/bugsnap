package main

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var secretKey = regexp.MustCompile(`(?i)(pass|secret|token|api[_-]?key|access[_-]?key|private[_-]?key|auth|credential|session|cookie|signature|salt)`)

var secretValue = regexp.MustCompile(`^(gh[pousr]_|github_pat_|sk-[A-Za-z0-9]|xox[abpr]-|AKIA|ASIA|glpat-|npm_|pypi-|eyJ)`)

const redacted = "[REDACTED]"

// redactEnv returns KEY=VALUE pairs with secret values masked, sorted by key.
func redactEnv(environ []string) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			out = append(out, kv)
			continue
		}
		if secretKey.MatchString(k) || secretValue.MatchString(v) {
			v = redacted
		}
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

// parseEnv turns redacted KEY=VALUE lines into a map.
func parseEnv(lines []string) map[string]string {
	m := make(map[string]string, len(lines))
	for _, kv := range lines {
		if k, v, ok := strings.Cut(kv, "="); ok {
			m[k] = v
		}
	}
	return m
}

// diffEnv describes how the current environment differs from the bundle's.
func diffEnv(bundle, current map[string]string) string {
	var b strings.Builder
	for _, k := range sortedKeys(bundle) {
		cv, ok := current[k]
		if !ok {
			b.WriteString("  - " + k + " (missing now)\n")
		} else if cv != bundle[k] {
			b.WriteString("  ~ " + k + ": " + bundle[k] + " -> " + cv + "\n")
		}
	}
	for _, k := range sortedKeys(current) {
		if _, ok := bundle[k]; !ok {
			b.WriteString("  + " + k + " (new)\n")
		}
	}
	if b.Len() == 0 {
		return "  (no environment differences)\n"
	}
	return b.String()
}

// isSecretFile blocks well-known credential files from untracked capture.
func isSecretFile(p string) bool {
	base := strings.ToLower(filepath.Base(p))
	if base == ".env" || strings.HasPrefix(base, ".env.") {
		return true
	}
	switch base {
	case ".npmrc", ".netrc", ".pgpass", "credentials":
		return true
	}
	for _, ext := range []string{".pem", ".key", ".p12", ".pfx"} {
		if strings.HasSuffix(base, ext) {
			return true
		}
	}
	return strings.HasPrefix(base, "id_rsa") || strings.HasPrefix(base, "id_ed25519")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
