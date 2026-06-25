package db

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMigrateAppliesEmbeddedMigrationsOnce(t *testing.T) {
	database := openTestDatabase(t)

	applied, err := database.Migrate(context.Background())
	if err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	if got, want := applied, 1; got != want {
		t.Fatalf("applied migrations = %d, want %d", got, want)
	}

	var version int64
	var name string
	var appliedAtMilliseconds int64
	if err := database.sql.QueryRow(
		"SELECT version, name, applied_at_ms FROM schema_migrations",
	).Scan(&version, &name, &appliedAtMilliseconds); err != nil {
		t.Fatalf("read applied migration: %v", err)
	}
	if version != 1 || name != "001_initial_schema.sql" {
		t.Fatalf("migration = (%d, %q), want (1, %q)", version, name, "001_initial_schema.sql")
	}
	if appliedAtMilliseconds <= 0 {
		t.Fatalf("applied_at_ms = %d, want positive timestamp", appliedAtMilliseconds)
	}

	applied, err = database.Migrate(context.Background())
	if err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}
	if applied != 0 {
		t.Fatalf("second applied migrations = %d, want 0", applied)
	}
}

func TestMigrateRollsBackFailedMigration(t *testing.T) {
	database := openTestDatabase(t)
	migrations := fstest.MapFS{
		"migrations/001_create_stable.sql": {
			Data: []byte("CREATE TABLE stable (id INTEGER PRIMARY KEY);"),
		},
		"migrations/002_fail.sql": {
			Data: []byte("CREATE TABLE rolled_back (id INTEGER); THIS IS NOT VALID SQL;"),
		},
	}

	applied, err := database.migrate(context.Background(), migrations)
	if err == nil {
		t.Fatal("migrate returned nil error, want failed migration error")
	}
	if applied != 1 {
		t.Fatalf("applied migrations = %d, want 1 before failure", applied)
	}
	if !tableExists(t, database, "stable") {
		t.Fatal("stable table does not exist after successful first migration")
	}
	if tableExists(t, database, "rolled_back") {
		t.Fatal("rolled_back table exists after failed migration")
	}

	var count int
	if err := database.sql.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("recorded migrations = %d, want 1", count)
	}
}

func TestMigrateRejectsUnknownAppliedMigration(t *testing.T) {
	database := openTestDatabase(t)
	knownMigrations := fstest.MapFS{
		"migrations/001_known.sql": {Data: []byte("CREATE TABLE known (id INTEGER);")},
	}
	if _, err := database.migrate(context.Background(), knownMigrations); err != nil {
		t.Fatalf("initial migrate returned error: %v", err)
	}

	_, err := database.migrate(context.Background(), fstest.MapFS{})
	if err == nil {
		t.Fatal("migrate returned nil error, want unknown migration error")
	}
	if !strings.Contains(err.Error(), "unknown migration version 1") {
		t.Fatalf("error = %q, want unknown migration message", err)
	}
}

func TestMigrateRejectsRenamedAppliedMigration(t *testing.T) {
	database := openTestDatabase(t)
	originalMigrations := fstest.MapFS{
		"migrations/001_original.sql": {Data: []byte("CREATE TABLE original (id INTEGER);")},
	}
	if _, err := database.migrate(context.Background(), originalMigrations); err != nil {
		t.Fatalf("initial migrate returned error: %v", err)
	}

	renamedMigrations := fstest.MapFS{
		"migrations/001_renamed.sql": {Data: []byte("SELECT 1;")},
	}
	_, err := database.migrate(context.Background(), renamedMigrations)
	if err == nil {
		t.Fatal("migrate returned nil error, want renamed migration error")
	}
	if !strings.Contains(err.Error(), "this binary provides") {
		t.Fatalf("error = %q, want migration name mismatch", err)
	}
}

func TestLoadMigrationsSortsByVersion(t *testing.T) {
	migrationFS := fstest.MapFS{
		"migrations/010_tenth.sql":  {Data: []byte("SELECT 10;")},
		"migrations/002_second.sql": {Data: []byte("SELECT 2;")},
	}

	migrations, err := loadMigrations(migrationFS)
	if err != nil {
		t.Fatalf("loadMigrations returned error: %v", err)
	}
	if got, want := migrations[0].version, int64(2); got != want {
		t.Fatalf("first migration version = %d, want %d", got, want)
	}
	if got, want := migrations[1].version, int64(10); got != want {
		t.Fatalf("second migration version = %d, want %d", got, want)
	}
}

func TestLoadMigrationsRejectsInvalidFiles(t *testing.T) {
	tests := []struct {
		name        string
		migrationFS fstest.MapFS
		want        string
	}{
		{
			name: "invalid name",
			migrationFS: fstest.MapFS{
				"migrations/initial.sql": {Data: []byte("SELECT 1;")},
			},
			want: "must be named",
		},
		{
			name: "duplicate version",
			migrationFS: fstest.MapFS{
				"migrations/001_first.sql":  {Data: []byte("SELECT 1;")},
				"migrations/001_second.sql": {Data: []byte("SELECT 2;")},
			},
			want: "version 1 is used",
		},
		{
			name: "empty migration",
			migrationFS: fstest.MapFS{
				"migrations/001_empty.sql": {Data: []byte(" \n")},
			},
			want: "is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadMigrations(tt.migrationFS)
			if err == nil {
				t.Fatal("loadMigrations returned nil error, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want text containing %q", err, tt.want)
			}
		})
	}
}
