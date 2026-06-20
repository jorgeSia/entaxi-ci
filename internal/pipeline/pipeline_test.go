package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSetsPipelineDirectory(t *testing.T) {
	dir := t.TempDir()
	config := []byte("steps:\n  - name: Test\n    command: go test ./...\n")
	if err := os.WriteFile(filepath.Join(dir, FileName), config, 0o600); err != nil {
		t.Fatalf("write pipeline config: %v", err)
	}

	pipeline, err := New(dir)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if pipeline.Dir != dir {
		t.Fatalf("pipeline directory = %q, want %q", pipeline.Dir, dir)
	}
}

func TestValidateDefaultsStepName(t *testing.T) {
	pipeline := Pipeline{
		Steps: []Step{{Command: "go test ./..."}},
	}

	if err := pipeline.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if got, want := pipeline.Steps[0].Name, "Step 1"; got != want {
		t.Fatalf("step name = %q, want %q", got, want)
	}
}

func TestValidateRejectsMissingCommand(t *testing.T) {
	pipeline := Pipeline{
		Steps: []Step{{Name: "Test"}},
	}

	if err := pipeline.Validate(); err == nil {
		t.Fatal("Validate returned nil error, want missing command error")
	}
}
