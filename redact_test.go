package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestRedactEnv(t *testing.T) {
	in := []string{
		"API_KEY=abc123",
		"DEBUG=1",
		"GITHUB_TOKEN=ghp_secret",
		"SAFE=ghp_abc",
		"PATH=/usr/bin",
	}
	got := redactEnv(in)
	want := []string{
		"API_KEY=[REDACTED]",
		"DEBUG=1",
		"GITHUB_TOKEN=[REDACTED]",
		"PATH=/usr/bin",
		"SAFE=[REDACTED]",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("redactEnv mismatch\n got: %v\nwant: %v", got, want)
	}
}

func TestIsSecretFile(t *testing.T) {
	for _, p := range []string{".env", ".env.local", "keys/server.pem", "id_rsa", ".npmrc"} {
		if !isSecretFile(p) {
			t.Errorf("%q should be treated as secret", p)
		}
	}
	for _, p := range []string{"main.go", "README.md", "config.yaml", "package.json"} {
		if isSecretFile(p) {
			t.Errorf("%q should not be treated as secret", p)
		}
	}
}

func TestDiffEnv(t *testing.T) {
	bundle := map[string]string{"A": "1", "B": "2"}
	current := map[string]string{"A": "1", "C": "3"}
	got := diffEnv(bundle, current)
	for _, want := range []string{"- B", "+ C"} {
		if !strings.Contains(got, want) {
			t.Fatalf("diffEnv missing %q in:\n%s", want, got)
		}
	}
}
