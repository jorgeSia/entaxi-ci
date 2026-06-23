// Package filemanager provides filesystem and path operations used by Entaxi.
package filemanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveDirectory converts path to an absolute path and verifies it is a directory.
func ResolveDirectory(path string) (string, error) {
	// Store absolute paths so later working-directory changes cannot affect them.
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", path, err)
	}

	// Check that the path exists and represents a directory.
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", absPath, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", absPath)
	}

	return absPath, nil
}

// Join combines path elements using the current operating system's separator.
func Join(elements ...string) string {
	return filepath.Join(elements...)
}

// ReadFile reads and returns all data from path.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}

	return data, nil
}

// EnsureDirectory creates path and any missing parents without changing existing permissions.
func EnsureDirectory(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("create directory %q: %w", path, err)
	}

	return nil
}

// WriteFileExclusive creates path only when it does not already exist.
func WriteFileExclusive(path string, data []byte, perm os.FileMode) (bool, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if errors.Is(err, os.ErrExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("create file %q: %w", path, err)
	}

	// Remove a partially written file so a later attempt can safely try again.
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return false, fmt.Errorf("write file %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return false, fmt.Errorf("close file %q: %w", path, err)
	}

	return true, nil
}

// ResolveUserPath expands a leading ~/ and requires an absolute result.
func ResolveUserPath(path string) (string, error) {
	switch {
	case path == "~":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		path = homeDir
	case strings.HasPrefix(path, "~/"):
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		path = filepath.Join(homeDir, strings.TrimPrefix(path, "~/"))
	}

	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("path %q must be absolute or start with ~/", path)
	}

	return filepath.Clean(path), nil
}
