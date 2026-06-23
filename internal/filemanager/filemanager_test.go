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

func TestEnsureDirectoryCreatesPrivateDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "entaxi")

	if err := EnsureDirectory(path, 0o700); err != nil {
		t.Fatalf("EnsureDirectory returned error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat created directory: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o700); got != want {
		t.Fatalf("directory permissions = %o, want %o", got, want)
	}
}

func TestEnsureDirectoryPreservesExistingPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("create existing directory: %v", err)
	}

	if err := EnsureDirectory(path, 0o700); err != nil {
		t.Fatalf("EnsureDirectory returned error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat existing directory: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o755); got != want {
		t.Fatalf("directory permissions = %o, want preserved %o", got, want)
	}
}

func TestWriteFileExclusiveCreatesPrivateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	created, err := WriteFileExclusive(path, []byte("first"), 0o600)
	if err != nil {
		t.Fatalf("WriteFileExclusive returned error: %v", err)
	}
	if !created {
		t.Fatal("created = false, want true")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat created file: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("file permissions = %o, want %o", got, want)
	}
}

func TestWriteFileExclusivePreservesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("original"), 0o600); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	created, err := WriteFileExclusive(path, []byte("replacement"), 0o600)
	if err != nil {
		t.Fatalf("WriteFileExclusive returned error: %v", err)
	}
	if created {
		t.Fatal("created = true, want false")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read existing file: %v", err)
	}
	if got, want := string(data), "original"; got != want {
		t.Fatalf("file contents = %q, want %q", got, want)
	}
}

func TestResolveUserPath(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{name: "absolute", path: filepath.Join(homeDir, "data"), want: filepath.Join(homeDir, "data")},
		{name: "home", path: "~", want: homeDir},
		{name: "home child", path: "~/.local/share/entaxi", want: filepath.Join(homeDir, ".local", "share", "entaxi")},
		{name: "relative", path: "data/entaxi", wantErr: true},
		{name: "named home", path: "~another/entaxi", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveUserPath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ResolveUserPath returned %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveUserPath returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("resolved path = %q, want %q", got, tt.want)
			}
		})
	}
}
