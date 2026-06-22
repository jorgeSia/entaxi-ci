package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "help command", args: []string{"help"}, want: rootUsage},
		{name: "short flag", args: []string{"-h"}, want: rootUsage},
		{name: "long flag", args: []string{"--help"}, want: rootUsage},
		{name: "run topic", args: []string{"help", "run"}, want: runUsage},
		{name: "run short flag", args: []string{"run", "-h"}, want: runUsage},
		{name: "run long flag", args: []string{"run", "--help"}, want: runUsage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var stderr bytes.Buffer
			cli := New(&out, &stderr)

			if code := cli.Run(tt.args); code != exitSuccess {
				t.Fatalf("exit code = %d, want %d", code, exitSuccess)
			}
			if got := out.String(); got != tt.want {
				t.Fatalf("standard output = %q, want %q", got, tt.want)
			}
			if stderr.Len() != 0 {
				t.Fatalf("standard error = %q, want empty", stderr.String())
			}
		})
	}
}

func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing command", want: "Usage:"},
		{name: "unknown command", args: []string{"unknown"}, want: "unknown command"},
		{name: "unknown help topic", args: []string{"help", "unknown"}, want: "unknown help topic"},
		{name: "too many help arguments", args: []string{"help", "run", "extra"}, want: "help expects at most one command"},
		{name: "too many run arguments", args: []string{"run", ".", "extra"}, want: "run expects at most one path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			var stderr bytes.Buffer
			cli := New(&out, &stderr)

			if code := cli.Run(tt.args); code != exitUsage {
				t.Fatalf("exit code = %d, want %d", code, exitUsage)
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("standard error = %q, want text containing %q", stderr.String(), tt.want)
			}
		})
	}
}

func TestRunPassingPipeline(t *testing.T) {
	dir := writePipeline(t, "printf pipeline-passed")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cli := New(&out, &stderr)

	if code := cli.Run([]string{"run", dir}); code != exitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", code, exitSuccess, stderr.String())
	}
	if !strings.Contains(out.String(), "pipeline-passed") {
		t.Fatalf("standard output = %q, want command output", out.String())
	}
}

func TestRunFailingPipeline(t *testing.T) {
	dir := writePipeline(t, "exit 7")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cli := New(&out, &stderr)

	if code := cli.Run([]string{"run", dir}); code != exitError {
		t.Fatalf("exit code = %d, want %d", code, exitError)
	}
	if !strings.Contains(out.String(), "Build failed") {
		t.Fatalf("standard output = %q, want failed build result", out.String())
	}
}

func writePipeline(t *testing.T, command string) string {
	t.Helper()

	dir := t.TempDir()
	config := []byte("steps:\n  - name: Test\n    command: " + command + "\n")
	if err := os.WriteFile(filepath.Join(dir, ".entaxi.yaml"), config, 0o600); err != nil {
		t.Fatalf("write pipeline config: %v", err)
	}

	return dir
}
