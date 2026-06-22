package filemanager

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDirectoryReturnsAbsolutePath(t *testing.T) {
	dir := t.TempDir()

	resolved, err := ResolveDirectory(dir)
	if err != nil {
		t.Fatalf("ResolveDirectory returned error: %v", err)
	}

	if !filepath.IsAbs(resolved) {
		t.Fatalf("resolved directory = %q, want an absolute path", resolved)
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "example.txt")
	if err := os.WriteFile(path, []byte("entaxi"), 0o600); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	data, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if got, want := string(data), "entaxi"; got != want {
		t.Fatalf("file contents = %q, want %q", got, want)
	}
}
