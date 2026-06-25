// Package db owns Entaxi's SQLite connection and schema migrations.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jorgeSia/entaxi-ci/internal/filemanager"
	_ "github.com/mattn/go-sqlite3"
)

const (
	driverName         = "sqlite3"
	privateFileMode    = 0o600
	busyTimeout        = 5 * time.Second
	maxOpenConnections = 1
)

// DB wraps the SQLite connection used by Entaxi.
type DB struct {
	sql  *sql.DB
	path string
}

// Open creates or opens an SQLite database and verifies the connection.
func Open(ctx context.Context, path string) (*DB, error) {
	resolvedPath, err := filemanager.ResolveUserPath(path)
	if err != nil {
		return nil, fmt.Errorf("resolve database path: %w", err)
	}

	// Pre-create the file so a new database receives private permissions.
	if _, err := filemanager.WriteFileExclusive(resolvedPath, nil, privateFileMode); err != nil {
		return nil, err
	}

	database, err := sql.Open(driverName, dataSourceName(resolvedPath))
	if err != nil {
		return nil, fmt.Errorf("open database %q: %w", resolvedPath, err)
	}

	// A single connection avoids SQLite lock contention in the initial version.
	database.SetMaxOpenConns(maxOpenConnections)
	database.SetMaxIdleConns(maxOpenConnections)

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("connect to database %q: %w", resolvedPath, err)
	}

	return &DB{sql: database, path: resolvedPath}, nil
}

// Path returns the absolute path of the open database.
func (d *DB) Path() string {
	return d.path
}

// Close releases the underlying database connection.
func (d *DB) Close() error {
	if err := d.sql.Close(); err != nil {
		return fmt.Errorf("close database %q: %w", d.path, err)
	}
	return nil
}

func dataSourceName(path string) string {
	query := url.Values{}
	query.Set("_busy_timeout", strconv.FormatInt(busyTimeout.Milliseconds(), 10))
	query.Set("_foreign_keys", "on")

	return (&url.URL{
		Scheme:   "file",
		Path:     path,
		RawQuery: query.Encode(),
	}).String()
}
