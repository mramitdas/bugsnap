package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestBundleRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "b.zip")

	m := Manifest{Version: 1, CreatedAt: "now", Files: []string{"git/status.txt"}}
	mb, _ := json.Marshal(m)
	files := map[string][]byte{
		"manifest.json":  mb,
		"git/status.txt": []byte("M file.go\n"),
	}
	if err := writeBundle(path, files); err != nil {
		t.Fatal(err)
	}

	got, gotFiles, err := readBundle(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 1 || got.CreatedAt != "now" {
		t.Fatalf("manifest mismatch: %+v", got)
	}
	if string(gotFiles["git/status.txt"]) != "M file.go\n" {
		t.Fatalf("file mismatch: %q", gotFiles["git/status.txt"])
	}
}

func TestSafeWriteRejectsTraversal(t *testing.T) {
	if err := safeWrite(t.TempDir(), "../evil", []byte("x")); err == nil {
		t.Fatal("expected traversal to be rejected")
	}
}
