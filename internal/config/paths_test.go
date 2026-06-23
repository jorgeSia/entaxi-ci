package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultPathsUsesXDGDirectories(t *testing.T) {
	homeDir := t.TempDir()
	configRoot := filepath.Join(homeDir, "xdg-config")
	dataRoot := filepath.Join(homeDir, "xdg-data")
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("XDG_DATA_HOME", dataRoot)

	paths, err := DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths returned error: %v", err)
	}

	assertPaths(t, paths, filepath.Join(configRoot, "entaxi"), filepath.Join(dataRoot, "entaxi"))
}

func TestDefaultPathsUsesHomeFallbacks(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")

	paths, err := DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths returned error: %v", err)
	}

	assertPaths(
		t,
		paths,
		filepath.Join(homeDir, ".config", "entaxi"),
		filepath.Join(homeDir, ".local", "share", "entaxi"),
	)
}

func TestDefaultPathsIgnoresRelativeXDGDirectories(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", "relative-config")
	t.Setenv("XDG_DATA_HOME", "relative-data")

	paths, err := DefaultPaths()
	if err != nil {
		t.Fatalf("DefaultPaths returned error: %v", err)
	}

	assertPaths(
		t,
		paths,
		filepath.Join(homeDir, ".config", "entaxi"),
		filepath.Join(homeDir, ".local", "share", "entaxi"),
	)
}

func TestNewPathsRejectsRelativeRoot(t *testing.T) {
	_, err := NewPaths("relative-config", t.TempDir())
	if err == nil {
		t.Fatal("NewPaths returned nil error, want relative path error")
	}
}

func TestPathsWithDataDirUpdatesDerivedPaths(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	paths, err := NewPaths(filepath.Join(homeDir, "config"), filepath.Join(homeDir, "data"))
	if err != nil {
		t.Fatalf("NewPaths returned error: %v", err)
	}

	updated, err := paths.WithDataDir("~/custom-entaxi-data")
	if err != nil {
		t.Fatalf("WithDataDir returned error: %v", err)
	}

	wantDataDir := filepath.Join(homeDir, "custom-entaxi-data")
	if updated.ConfigDir != paths.ConfigDir || updated.ConfigFile != paths.ConfigFile {
		t.Fatal("WithDataDir changed configuration paths")
	}
	if updated.DataDir != wantDataDir {
		t.Fatalf("data directory = %q, want %q", updated.DataDir, wantDataDir)
	}
	if updated.DatabaseFile != filepath.Join(wantDataDir, DatabaseFileName) {
		t.Fatalf("database file = %q, want path below data directory", updated.DatabaseFile)
	}
	if updated.LogsDir != filepath.Join(wantDataDir, LogsDirName) {
		t.Fatalf("logs directory = %q, want path below data directory", updated.LogsDir)
	}
}

func assertPaths(t *testing.T, paths Paths, configDir, dataDir string) {
	t.Helper()

	if paths.ConfigDir != configDir {
		t.Fatalf("config directory = %q, want %q", paths.ConfigDir, configDir)
	}
	if paths.ConfigFile != filepath.Join(configDir, ConfigFileName) {
		t.Fatalf("config file = %q, want path below config directory", paths.ConfigFile)
	}
	if paths.DataDir != dataDir {
		t.Fatalf("data directory = %q, want %q", paths.DataDir, dataDir)
	}
	if paths.DatabaseFile != filepath.Join(dataDir, DatabaseFileName) {
		t.Fatalf("database file = %q, want path below data directory", paths.DatabaseFile)
	}
	if paths.LogsDir != filepath.Join(dataDir, LogsDirName) {
		t.Fatalf("logs directory = %q, want path below data directory", paths.LogsDir)
	}
}
