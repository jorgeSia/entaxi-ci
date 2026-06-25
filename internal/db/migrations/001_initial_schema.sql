CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE,
    default_branch TEXT,
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),
    created_at_ms INTEGER NOT NULL,
    updated_at_ms INTEGER NOT NULL
);

CREATE TABLE builds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'passed', 'failed', 'canceled', 'error')),
    trigger TEXT NOT NULL CHECK (trigger IN ('manual', 'poll', 'git-hook')),
    commit_hash TEXT,
    branch TEXT,
    dirty_worktree INTEGER NOT NULL DEFAULT 0 CHECK (dirty_worktree IN (0, 1)),
    started_at_ms INTEGER,
    finished_at_ms INTEGER,
    duration_ms INTEGER CHECK (duration_ms IS NULL OR duration_ms >= 0),
    exit_code INTEGER,
    created_at_ms INTEGER NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE step_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    build_id INTEGER NOT NULL,
    step_index INTEGER NOT NULL CHECK (step_index >= 1),
    step_name TEXT NOT NULL,
    command TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'passed', 'failed', 'canceled', 'error')),
    started_at_ms INTEGER,
    finished_at_ms INTEGER,
    duration_ms INTEGER CHECK (duration_ms IS NULL OR duration_ms >= 0),
    exit_code INTEGER,
    log_path TEXT,
    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE
);

CREATE INDEX idx_builds_project_created_at
    ON builds(project_id, created_at_ms DESC);

CREATE INDEX idx_builds_status_created_at
    ON builds(status, created_at_ms);

CREATE UNIQUE INDEX idx_step_runs_build_step
    ON step_runs(build_id, step_index);
