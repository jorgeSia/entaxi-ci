package db

import (
	"context"
	"fmt"
	"testing"
)

func TestInitialSchemaContainsExpectedTablesAndColumns(t *testing.T) {
	database := migratedTestDatabase(t)

	expectedColumns := map[string][]string{
		"projects": {
			"id", "name", "path", "default_branch", "enabled", "created_at_ms", "updated_at_ms",
		},
		"builds": {
			"id", "project_id", "status", "trigger", "commit_hash", "branch", "dirty_worktree",
			"started_at_ms", "finished_at_ms", "duration_ms", "exit_code", "created_at_ms",
		},
		"step_runs": {
			"id", "build_id", "step_index", "step_name", "command", "status", "started_at_ms",
			"finished_at_ms", "duration_ms", "exit_code", "log_path",
		},
	}

	for table, columns := range expectedColumns {
		if !tableExists(t, database, table) {
			t.Fatalf("table %q does not exist", table)
		}
		got := tableColumns(t, database, table)
		for _, column := range columns {
			if !got[column] {
				t.Errorf("table %q is missing column %q", table, column)
			}
		}
		if len(got) != len(columns) {
			t.Errorf("table %q has %d columns, want %d", table, len(got), len(columns))
		}
	}
}

func TestInitialSchemaContainsExpectedIndexes(t *testing.T) {
	database := migratedTestDatabase(t)

	for _, index := range []string{
		"idx_builds_project_created_at",
		"idx_builds_status_created_at",
		"idx_step_runs_build_step",
	} {
		var count int
		if err := database.sql.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name = ?",
			index,
		).Scan(&count); err != nil {
			t.Fatalf("query index %q: %v", index, err)
		}
		if count != 1 {
			t.Errorf("index %q count = %d, want 1", index, count)
		}
	}
}

func TestInitialSchemaCascadesProjectDeletion(t *testing.T) {
	database := migratedTestDatabase(t)
	ctx := context.Background()

	projectResult, err := database.sql.ExecContext(ctx, `
        INSERT INTO projects (name, path, enabled, created_at_ms, updated_at_ms)
        VALUES ('entaxi', '/tmp/entaxi', 1, 1, 1)`)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatalf("read project ID: %v", err)
	}

	buildResult, err := database.sql.ExecContext(ctx, `
        INSERT INTO builds (project_id, status, trigger, created_at_ms)
        VALUES (?, 'passed', 'manual', 1)`, projectID)
	if err != nil {
		t.Fatalf("insert build: %v", err)
	}
	buildID, err := buildResult.LastInsertId()
	if err != nil {
		t.Fatalf("read build ID: %v", err)
	}

	if _, err := database.sql.ExecContext(ctx, `
        INSERT INTO step_runs (build_id, step_index, step_name, command, status)
        VALUES (?, 1, 'Test', 'go test ./...', 'passed')`, buildID); err != nil {
		t.Fatalf("insert step run: %v", err)
	}

	if _, err := database.sql.ExecContext(ctx, "DELETE FROM projects WHERE id = ?", projectID); err != nil {
		t.Fatalf("delete project: %v", err)
	}

	for _, table := range []string{"builds", "step_runs"} {
		var count int
		if err := database.sql.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Errorf("%s rows = %d, want 0 after cascade", table, count)
		}
	}
}

func TestInitialSchemaEnforcesStatusAndStepIndex(t *testing.T) {
	database := migratedTestDatabase(t)
	ctx := context.Background()

	projectResult, err := database.sql.ExecContext(ctx, `
        INSERT INTO projects (name, path, created_at_ms, updated_at_ms)
        VALUES ('entaxi', '/tmp/entaxi', 1, 1)`)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	projectID, err := projectResult.LastInsertId()
	if err != nil {
		t.Fatalf("read project ID: %v", err)
	}

	if _, err := database.sql.ExecContext(ctx, `
        INSERT INTO builds (project_id, status, trigger, created_at_ms)
        VALUES (?, 'unknown', 'manual', 1)`, projectID); err == nil {
		t.Fatal("invalid build status was accepted")
	}

	buildResult, err := database.sql.ExecContext(ctx, `
        INSERT INTO builds (project_id, status, trigger, created_at_ms)
        VALUES (?, 'running', 'manual', 1)`, projectID)
	if err != nil {
		t.Fatalf("insert build: %v", err)
	}
	buildID, err := buildResult.LastInsertId()
	if err != nil {
		t.Fatalf("read build ID: %v", err)
	}

	if _, err := database.sql.ExecContext(ctx, `
        INSERT INTO step_runs (build_id, step_index, step_name, command, status)
        VALUES (?, 0, 'Test', 'go test ./...', 'running')`, buildID); err == nil {
		t.Fatal("invalid step index was accepted")
	}
}

func migratedTestDatabase(t *testing.T) *DB {
	t.Helper()

	database := openTestDatabase(t)
	if _, err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	return database
}

func tableExists(t *testing.T, database *DB, table string) bool {
	t.Helper()

	var count int
	if err := database.sql.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
		table,
	).Scan(&count); err != nil {
		t.Fatalf("query table %q: %v", table, err)
	}
	return count == 1
}

func tableColumns(t *testing.T, database *DB, table string) map[string]bool {
	t.Helper()

	rows, err := database.sql.Query(fmt.Sprintf("PRAGMA table_info(%q)", table))
	if err != nil {
		t.Fatalf("query columns for %q: %v", table, err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var id int
		var name string
		var columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := rows.Scan(&id, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan column for %q: %v", table, err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read columns for %q: %v", table, err)
	}
	return columns
}
