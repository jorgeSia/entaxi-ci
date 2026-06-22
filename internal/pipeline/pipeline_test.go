package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSetsAbsolutePipelineDirectory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "project")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("create project directory: %v", err)
	}
	writePipelineConfig(t, dir, "steps:\n  - name: Test\n    command: go test ./...\n")

	t.Chdir(parent)
	pipeline, err := New("project")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if pipeline.Dir != dir {
		t.Fatalf("pipeline directory = %q, want %q", pipeline.Dir, dir)
	}
}

func TestNewRejectsMissingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")

	_, err := New(dir)
	if err == nil {
		t.Fatal("New returned nil error, want missing directory error")
	}
	if !strings.Contains(err.Error(), dir) {
		t.Fatalf("error = %q, want path %q", err, dir)
	}
}

func TestNewRejectsFileAsDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "project")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("write project file: %v", err)
	}

	_, err := New(path)
	if err == nil {
		t.Fatal("New returned nil error, want non-directory error")
	}
	if !strings.Contains(err.Error(), "is not a directory") {
		t.Fatalf("error = %q, want non-directory message", err)
	}
}

func TestNewRejectsMissingConfiguration(t *testing.T) {
	dir := t.TempDir()

	_, err := New(dir)
	if err == nil {
		t.Fatal("New returned nil error, want missing configuration error")
	}
	if !strings.Contains(err.Error(), filepath.Join(dir, FileName)) {
		t.Fatalf("error = %q, want configuration path", err)
	}
}

func TestNewIncludesConfigurationPathInParseError(t *testing.T) {
	dir := t.TempDir()
	writePipelineConfig(t, dir, "steps: [\n")

	_, err := New(dir)
	if err == nil {
		t.Fatal("New returned nil error, want parse error")
	}
	if !strings.Contains(err.Error(), filepath.Join(dir, FileName)) {
		t.Fatalf("error = %q, want configuration path", err)
	}
}

func TestValidateNormalizesPipeline(t *testing.T) {
	pipeline := Pipeline{
		Name: "  Entaxi  ",
		Steps: []Step{
			{Name: "  Test  ", Command: "  go test ./...  "},
			{Command: "go build ./..."},
		},
	}

	if err := pipeline.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if got, want := pipeline.Name, "Entaxi"; got != want {
		t.Fatalf("pipeline name = %q, want %q", got, want)
	}
	if got, want := pipeline.Steps[0].Name, "Test"; got != want {
		t.Fatalf("first step name = %q, want %q", got, want)
	}
	if got, want := pipeline.Steps[0].Command, "go test ./..."; got != want {
		t.Fatalf("first step command = %q, want %q", got, want)
	}
	if got, want := pipeline.Steps[1].Name, "Step 2"; got != want {
		t.Fatalf("second step name = %q, want %q", got, want)
	}
}

func TestValidateRejectsInvalidPipeline(t *testing.T) {
	tests := []struct {
		name     string
		pipeline Pipeline
		want     string
	}{
		{
			name: "no steps",
			want: "must define at least one step",
		},
		{
			name:     "missing command",
			pipeline: Pipeline{Steps: []Step{{Name: "Test"}}},
			want:     "must define a command",
		},
		{
			name:     "whitespace command",
			pipeline: Pipeline{Steps: []Step{{Name: "Test", Command: "  \t "}}},
			want:     "must define a command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pipeline.Validate()
			if err == nil {
				t.Fatal("Validate returned nil error, want validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want text containing %q", err, tt.want)
			}
		})
	}
}

func writePipelineConfig(t *testing.T, dir, config string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(config), 0o600); err != nil {
		t.Fatalf("write pipeline config: %v", err)
	}
}
