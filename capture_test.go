package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureDepsWalksMonorepo(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		full := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("package.json", "{}")
	write("packages/web/package.json", "{}")
	write("packages/web/yarn.lock", "x")
	write("node_modules/dep/package.json", "{}")
	write("packages/web/src/index.ts", "x")

	got := captureDeps(root)
	for _, want := range []string{"deps/package.json", "deps/packages/web/package.json", "deps/packages/web/yarn.lock"} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing %s (got %v)", want, sortedKeys(got))
		}
	}
	if _, ok := got["deps/node_modules/dep/package.json"]; ok {
		t.Error("node_modules should be skipped")
	}
}
