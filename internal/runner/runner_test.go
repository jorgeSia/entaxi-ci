package runner

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/jorgeSia/entaxi-ci/internal/pipeline"
)

func TestRunStreamsOutputAndPasses(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	runner := New(&out, &stderr)
	pipeline := pipeline.Pipeline{
		Dir: t.TempDir(),
		Steps: []pipeline.Step{
			{Name: "Standard output", Command: "printf output"},
			{Name: "Standard error", Command: "printf warning >&2"},
		},
	}

	result, err := runner.Run(context.Background(), pipeline)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !result.Passed {
		t.Fatal("result did not pass")
	}
	if !strings.Contains(out.String(), "output") {
		t.Fatalf("standard output = %q, want command output", out.String())
	}
	if got, want := stderr.String(), "warning"; got != want {
		t.Fatalf("standard error = %q, want %q", got, want)
	}
}

func TestRunStopsAfterFailedStep(t *testing.T) {
	var out bytes.Buffer
	var stderr bytes.Buffer
	runner := New(&out, &stderr)
	pipeline := pipeline.Pipeline{
		Dir: t.TempDir(),
		Steps: []pipeline.Step{
			{Name: "Fail", Command: "exit 7"},
			{Name: "Should not run", Command: "printf should-not-run"},
		},
	}

	result, err := runner.Run(context.Background(), pipeline)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Passed {
		t.Fatal("result passed, want failure")
	}
	if got, want := result.ExitCode, 7; got != want {
		t.Fatalf("exit code = %d, want %d", got, want)
	}
	if strings.Contains(out.String(), "should-not-run") {
		t.Fatalf("standard output = %q, later step unexpectedly ran", out.String())
	}
}
