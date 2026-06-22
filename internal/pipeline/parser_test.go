package pipeline

import (
	"strings"
	"testing"
)

func TestParseValidConfiguration(t *testing.T) {
	data := []byte("name: Entaxi\nsteps:\n  - name: Test\n    command: go test ./...\n")

	pipeline, err := parse(data)
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}

	if got, want := pipeline.Name, "Entaxi"; got != want {
		t.Fatalf("pipeline name = %q, want %q", got, want)
	}
	if got, want := len(pipeline.Steps), 1; got != want {
		t.Fatalf("step count = %d, want %d", got, want)
	}
}

func TestParseRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "malformed YAML",
			data: "steps: [\n",
			want: "decode YAML",
		},
		{
			name: "unknown pipeline field",
			data: "stpes: []\n",
			want: "field stpes not found",
		},
		{
			name: "unknown step field",
			data: "steps:\n  - name: Test\n    commands: go test ./...\n",
			want: "field commands not found",
		},
		{
			name: "multiple documents",
			data: "steps: []\n---\nsteps: []\n",
			want: "exactly one YAML document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse([]byte(tt.data))
			if err == nil {
				t.Fatal("parse returned nil error, want parse error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want text containing %q", err, tt.want)
			}
		})
	}
}
