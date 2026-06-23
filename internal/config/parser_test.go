package config

import (
	"strings"
	"testing"
)

func TestParseValidConfig(t *testing.T) {
	data := []byte(`data_dir: /tmp/entaxi
host: 127.0.0.1
port: 7878
max_parallel_builds: 1
poll_interval_seconds: 30
`)

	config, err := parse(data)
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if got, want := config.DataDir, "/tmp/entaxi"; got != want {
		t.Fatalf("data directory = %q, want %q", got, want)
	}
}

func TestParseRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{name: "malformed YAML", data: "data_dir: [\n", want: "decode YAML"},
		{name: "unknown field", data: "data_directory: /tmp/entaxi\n", want: "field data_directory not found"},
		{name: "multiple documents", data: "data_dir: /tmp/one\n---\ndata_dir: /tmp/two\n", want: "exactly one YAML document"},
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
