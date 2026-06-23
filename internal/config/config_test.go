package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jorgeSia/entaxi-ci/internal/filemanager"
)

func TestDefaultConfig(t *testing.T) {
	paths, err := NewPaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("NewPaths returned error: %v", err)
	}

	config := Default(paths)

	if config.DataDir != paths.DataDir {
		t.Fatalf("data directory = %q, want %q", config.DataDir, paths.DataDir)
	}
	if config.Host != DefaultHost {
		t.Fatalf("host = %q, want %q", config.Host, DefaultHost)
	}
	if config.Port != DefaultPort {
		t.Fatalf("port = %d, want %d", config.Port, DefaultPort)
	}
	if config.MaxParallelBuilds != DefaultMaxParallelBuilds {
		t.Fatalf("max parallel builds = %d, want %d", config.MaxParallelBuilds, DefaultMaxParallelBuilds)
	}
	if config.PollIntervalSeconds != DefaultPollIntervalSeconds {
		t.Fatalf("poll interval = %d, want %d", config.PollIntervalSeconds, DefaultPollIntervalSeconds)
	}
}

func TestLoadNormalizesConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	path := filepath.Join(t.TempDir(), ConfigFileName)
	writeConfig(t, path, `data_dir: ~/.local/share/custom-entaxi
host: " 0.0.0.0 "
port: 8080
max_parallel_builds: 2
poll_interval_seconds: 60
`)

	config, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got, want := config.DataDir, filepath.Join(homeDir, ".local", "share", "custom-entaxi"); got != want {
		t.Fatalf("data directory = %q, want %q", got, want)
	}
	if got, want := config.Host, "0.0.0.0"; got != want {
		t.Fatalf("host = %q, want %q", got, want)
	}
}

func TestLoadIncludesConfigPathInParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), ConfigFileName)
	writeConfig(t, path, "data_dir: [\n")

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load returned nil error, want parse error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error = %q, want configuration path %q", err, path)
	}
}

func TestCreateConfigUsesPrivatePermissions(t *testing.T) {
	configRoot := t.TempDir()
	paths, err := NewPaths(configRoot, t.TempDir())
	if err != nil {
		t.Fatalf("NewPaths returned error: %v", err)
	}
	if err := filemanager.EnsureDirectory(paths.ConfigDir, 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	created, err := Create(paths.ConfigFile, Default(paths))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if !created {
		t.Fatal("created = false, want true")
	}

	info, err := os.Stat(paths.ConfigFile)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("config permissions = %o, want %o", got, want)
	}

	loaded, err := Load(paths.ConfigFile)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded != Default(paths) {
		t.Fatalf("loaded config = %#v, want %#v", loaded, Default(paths))
	}
}

func TestCreateConfigPreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFileName)
	original := []byte("personal configuration")
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	config := Config{
		DataDir:             filepath.Join(dir, "data"),
		Host:                DefaultHost,
		Port:                DefaultPort,
		MaxParallelBuilds:   DefaultMaxParallelBuilds,
		PollIntervalSeconds: DefaultPollIntervalSeconds,
	}
	created, err := Create(path, config)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created {
		t.Fatal("created = true, want false")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read existing config: %v", err)
	}
	if string(data) != string(original) {
		t.Fatalf("config contents = %q, want preserved %q", data, original)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat existing config: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("config permissions = %o, want preserved %o", got, want)
	}
}

func TestValidateRejectsInvalidConfig(t *testing.T) {
	valid := Config{
		DataDir:             filepath.Join(t.TempDir(), "data"),
		Host:                DefaultHost,
		Port:                DefaultPort,
		MaxParallelBuilds:   DefaultMaxParallelBuilds,
		PollIntervalSeconds: DefaultPollIntervalSeconds,
	}

	tests := []struct {
		name   string
		change func(*Config)
		want   string
	}{
		{name: "empty data directory", change: func(c *Config) { c.DataDir = "" }, want: "data_dir must not be empty"},
		{name: "relative data directory", change: func(c *Config) { c.DataDir = "relative/data" }, want: "must be absolute"},
		{name: "empty host", change: func(c *Config) { c.Host = "  " }, want: "host must not be empty"},
		{name: "zero port", change: func(c *Config) { c.Port = 0 }, want: "port must be between"},
		{name: "large port", change: func(c *Config) { c.Port = 65536 }, want: "port must be between"},
		{name: "zero parallel builds", change: func(c *Config) { c.MaxParallelBuilds = 0 }, want: "max_parallel_builds"},
		{name: "zero poll interval", change: func(c *Config) { c.PollIntervalSeconds = 0 }, want: "poll_interval_seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := valid
			tt.change(&config)

			err := config.Validate()
			if err == nil {
				t.Fatal("Validate returned nil error, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want text containing %q", err, tt.want)
			}
		})
	}
}

func writeConfig(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}
