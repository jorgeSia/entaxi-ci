// Package config defines Entaxi's global configuration and filesystem paths.
package config

import (
	"fmt"
	"strings"

	"github.com/jorgeSia/entaxi-ci/internal/filemanager"
	"gopkg.in/yaml.v3"
)

const (
	DefaultHost                = "127.0.0.1"
	DefaultPort                = 7878
	DefaultMaxParallelBuilds   = 1
	DefaultPollIntervalSeconds = 30
	privateFileMode            = 0o600
)

// Config contains settings shared by all Entaxi projects for one user.
type Config struct {
	// DataDir contains Entaxi's database and build logs.
	DataDir string `yaml:"data_dir"`
	// Host is the address used by the future HTTP server.
	Host string `yaml:"host"`
	// Port is the port used by the future HTTP server.
	Port int `yaml:"port"`
	// MaxParallelBuilds limits how many pipelines may run concurrently.
	MaxParallelBuilds int `yaml:"max_parallel_builds"`
	// PollIntervalSeconds controls how often repositories will be checked for changes.
	PollIntervalSeconds int `yaml:"poll_interval_seconds"`
}

// Default creates a global configuration using the default data directory.
func Default(paths Paths) Config {
	return Config{
		DataDir:             paths.DataDir,
		Host:                DefaultHost,
		Port:                DefaultPort,
		MaxParallelBuilds:   DefaultMaxParallelBuilds,
		PollIntervalSeconds: DefaultPollIntervalSeconds,
	}
}

// Load reads, parses, and validates an existing global configuration.
func Load(path string) (Config, error) {
	data, err := filemanager.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	config, err := parse(data)
	if err != nil {
		return Config{}, fmt.Errorf("parse global config %q: %w", path, err)
	}
	if err := config.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate global config %q: %w", path, err)
	}

	return config, nil
}

// Create writes config with private permissions unless path already exists.
func Create(path string, config Config) (bool, error) {
	if err := config.Validate(); err != nil {
		return false, fmt.Errorf("validate global config: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return false, fmt.Errorf("encode global config: %w", err)
	}

	created, err := filemanager.WriteFileExclusive(path, data, privateFileMode)
	if err != nil {
		return false, err
	}
	return created, nil
}

// Validate normalizes configuration values and rejects unusable settings.
func (c *Config) Validate() error {
	c.DataDir = strings.TrimSpace(c.DataDir)
	c.Host = strings.TrimSpace(c.Host)

	if c.DataDir == "" {
		return fmt.Errorf("data_dir must not be empty")
	}
	resolvedDataDir, err := filemanager.ResolveUserPath(c.DataDir)
	if err != nil {
		return fmt.Errorf("invalid data_dir: %w", err)
	}
	c.DataDir = resolvedDataDir

	if c.Host == "" {
		return fmt.Errorf("host must not be empty")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.MaxParallelBuilds < 1 {
		return fmt.Errorf("max_parallel_builds must be at least 1")
	}
	if c.PollIntervalSeconds < 1 {
		return fmt.Errorf("poll_interval_seconds must be at least 1")
	}

	return nil
}
