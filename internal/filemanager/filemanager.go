// Package filemanager provides filesystem and path operations used by Entaxi.
package filemanager

import (
	"fmt"
	"os"
	"path/filepath"
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
