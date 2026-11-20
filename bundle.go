package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

const maxRead = 32 << 20

func writeBundle(path string, files map[string][]byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	for _, name := range sortedKeys(files) {
		w, err := zw.Create(name)
		if err != nil {
			f.Close()
			return err
		}
		if _, err := w.Write(files[name]); err != nil {
			f.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func readBundle(path string) (Manifest, map[string][]byte, error) {
	var m Manifest
	files := map[string][]byte{}

	zr, err := zip.OpenReader(path)
	if err != nil {
		return m, nil, err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return m, nil, err
		}
		b, err := io.ReadAll(io.LimitReader(rc, maxRead))
		rc.Close()
		if err != nil {
			return m, nil, err
		}
		files[f.Name] = b
	}
	if b, ok := files["manifest.json"]; ok {
		if err := json.Unmarshal(b, &m); err != nil {
			return m, nil, err
		}
	}
	return m, files, nil
}

func emitJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func sortedFiles(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		if name != "manifest.json" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
