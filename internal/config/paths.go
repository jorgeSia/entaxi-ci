package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jorgeSia/entaxi-ci/internal/filemanager"
)

const (
	applicationDirName = "entaxi"
	ConfigFileName     = "config.yaml"
	DatabaseFileName   = "entaxi.db"
	LogsDirName        = "logs"
)

// Paths contains every filesystem location used by one Entaxi installation.
type Paths struct {
	ConfigDir    string
	ConfigFile   string
	DataDir      string
	DatabaseFile string
	LogsDir      string
}

// DefaultPaths resolves per-user paths from XDG environment variables and home fallbacks.
func DefaultPaths() (Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve home directory: %w", err)
	}

	// XDG requires absolute values; invalid values are ignored in favor of defaults.
	configRoot := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(configRoot) {
		configRoot = filemanager.Join(homeDir, ".config")
	}

	dataRoot := os.Getenv("XDG_DATA_HOME")
	if !filepath.IsAbs(dataRoot) {
		dataRoot = filemanager.Join(homeDir, ".local", "share")
	}

	return NewPaths(configRoot, dataRoot)
}

// NewPaths creates Entaxi paths below explicit configuration and data roots.
func NewPaths(configRoot, dataRoot string) (Paths, error) {
	resolvedConfigRoot, err := filemanager.ResolveUserPath(configRoot)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve config root: %w", err)
	}

	resolvedDataRoot, err := filemanager.ResolveUserPath(dataRoot)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve data root: %w", err)
	}

	configDir := filemanager.Join(resolvedConfigRoot, applicationDirName)
	dataDir := filemanager.Join(resolvedDataRoot, applicationDirName)

	return Paths{
		ConfigDir:    configDir,
		ConfigFile:   filemanager.Join(configDir, ConfigFileName),
		DataDir:      dataDir,
		DatabaseFile: filemanager.Join(dataDir, DatabaseFileName),
		LogsDir:      filemanager.Join(dataDir, LogsDirName),
	}, nil
}

// WithDataDir returns a copy whose state paths use the configured data directory.
func (p Paths) WithDataDir(dataDir string) (Paths, error) {
	resolvedDataDir, err := filemanager.ResolveUserPath(dataDir)
	if err != nil {
		return Paths{}, fmt.Errorf("resolve data directory: %w", err)
	}

	p.DataDir = resolvedDataDir
	p.DatabaseFile = filemanager.Join(resolvedDataDir, DatabaseFileName)
	p.LogsDir = filemanager.Join(resolvedDataDir, LogsDirName)
	return p, nil
}
