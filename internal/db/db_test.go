package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesPrivateDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entaxi.db")

	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close returned error: %v", err)
		}
	})

	if got, want := database.Path(), path; got != want {
		t.Fatalf("database path = %q, want %q", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat database: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("database permissions = %o, want %o", got, want)
	}
}

func TestOpenPreservesExistingPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "entaxi.db")
	if err := os.WriteFile(path, nil, 0o640); err != nil {
		t.Fatalf("create existing database: %v", err)
	}

	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close returned error: %v", err)
		}
	})

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat database: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o640); got != want {
		t.Fatalf("database permissions = %o, want preserved %o", got, want)
	}
}

func TestOpenConfiguresSQLiteConnection(t *testing.T) {
	database := openTestDatabase(t)

	var foreignKeys int
	if err := database.sql.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign_keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}

	var timeoutMilliseconds int64
	if err := database.sql.QueryRow("PRAGMA busy_timeout").Scan(&timeoutMilliseconds); err != nil {
		t.Fatalf("read busy_timeout pragma: %v", err)
	}
	if got, want := timeoutMilliseconds, busyTimeout.Milliseconds(); got != want {
		t.Fatalf("busy_timeout = %d, want %d", got, want)
	}

	if got, want := database.sql.Stats().MaxOpenConnections, maxOpenConnections; got != want {
		t.Fatalf("max open connections = %d, want %d", got, want)
	}
}

func TestOpenRejectsMissingParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "entaxi.db")

	if _, err := Open(context.Background(), path); err == nil {
		t.Fatal("Open returned nil error, want missing parent error")
	}
}

func openTestDatabase(t *testing.T) *DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "entaxi.db")
	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close returned error: %v", err)
		}
	})
	return database
}
