package config

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// parse decodes one strict YAML document without performing filesystem work.
func parse(data []byte) (Config, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("decode YAML: %w", err)
	}

	var extraDocument any
	if err := decoder.Decode(&extraDocument); err != io.EOF {
		if err != nil {
			return Config{}, fmt.Errorf("decode YAML: %w", err)
		}
		return Config{}, fmt.Errorf("configuration must contain exactly one YAML document")
	}

	return config, nil
}
