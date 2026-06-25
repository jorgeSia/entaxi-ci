package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

const createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    applied_at_ms INTEGER NOT NULL
);`

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type migration struct {
	version int64
	name    string
	sql     string
}

// Migrate applies every embedded migration that has not already run.
func (d *DB) Migrate(ctx context.Context) (int, error) {
	return d.migrate(ctx, embeddedMigrations)
}

func (d *DB) migrate(ctx context.Context, migrationFS fs.FS) (int, error) {
	if _, err := d.sql.ExecContext(ctx, createMigrationsTableSQL); err != nil {
		return 0, fmt.Errorf("create schema_migrations table: %w", err)
	}

	migrations, err := loadMigrations(migrationFS)
	if err != nil {
		return 0, err
	}
	appliedMigrations, err := d.appliedMigrations(ctx)
	if err != nil {
		return 0, err
	}
	if err := validateAppliedMigrations(migrations, appliedMigrations); err != nil {
		return 0, err
	}

	appliedCount := 0
	for _, migration := range migrations {
		if _, exists := appliedMigrations[migration.version]; exists {
			continue
		}
		if err := d.applyMigration(ctx, migration); err != nil {
			return appliedCount, err
		}
		appliedCount++
	}

	return appliedCount, nil
}

func loadMigrations(migrationFS fs.FS) ([]migration, error) {
	paths, err := fs.Glob(migrationFS, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}

	migrations := make([]migration, 0, len(paths))
	versions := make(map[int64]string, len(paths))
	for _, path := range paths {
		migration, err := loadMigration(migrationFS, path)
		if err != nil {
			return nil, err
		}
		if previous, exists := versions[migration.version]; exists {
			return nil, fmt.Errorf("migration version %d is used by both %q and %q", migration.version, previous, migration.name)
		}
		versions[migration.version] = migration.name
		migrations = append(migrations, migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations, nil
}

func loadMigration(migrationFS fs.FS, migrationPath string) (migration, error) {
	name := path.Base(migrationPath)
	stem := strings.TrimSuffix(name, path.Ext(name))
	parts := strings.SplitN(stem, "_", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return migration{}, fmt.Errorf("migration %q must be named <version>_<name>.sql", name)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || version < 1 {
		return migration{}, fmt.Errorf("migration %q has invalid version %q", name, parts[0])
	}

	data, err := fs.ReadFile(migrationFS, migrationPath)
	if err != nil {
		return migration{}, fmt.Errorf("read migration %q: %w", name, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return migration{}, fmt.Errorf("migration %q is empty", name)
	}

	return migration{
		version: version,
		name:    name,
		sql:     string(data),
	}, nil
}

func (d *DB) appliedMigrations(ctx context.Context) (map[int64]string, error) {
	rows, err := d.sql.QueryContext(ctx, "SELECT version, name FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	migrations := make(map[int64]string)
	for rows.Next() {
		var version int64
		var name string
		if err := rows.Scan(&version, &name); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		migrations[version] = name
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}

	return migrations, nil
}

func validateAppliedMigrations(available []migration, applied map[int64]string) error {
	known := make(map[int64]string, len(available))
	for _, migration := range available {
		known[migration.version] = migration.name
	}

	for version, appliedName := range applied {
		availableName, exists := known[version]
		if !exists {
			return fmt.Errorf("database contains unknown migration version %d (%q)", version, appliedName)
		}
		if availableName != appliedName {
			return fmt.Errorf(
				"database migration version %d is named %q, but this binary provides %q",
				version,
				appliedName,
				availableName,
			)
		}
	}

	return nil
}

func (d *DB) applyMigration(ctx context.Context, pending migration) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %q: %w", pending.name, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, pending.sql); err != nil {
		return fmt.Errorf("apply migration %q: %w", pending.name, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		"INSERT INTO schema_migrations (version, name, applied_at_ms) VALUES (?, ?, ?)",
		pending.version,
		pending.name,
		time.Now().UTC().UnixMilli(),
	); err != nil {
		return fmt.Errorf("record migration %q: %w", pending.name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %q: %w", pending.name, err)
	}

	return nil
}
