package pipeline

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// parse decodes YAML data into a pipeline without performing filesystem work.
func parse(data []byte) (Pipeline, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	// Reject misspelled or unsupported fields instead of silently ignoring them.
	decoder.KnownFields(true)

	var pipeline Pipeline
	if err := decoder.Decode(&pipeline); err != nil {
		return Pipeline{}, fmt.Errorf("decode YAML: %w", err)
	}

	// A pipeline file represents exactly one configuration document.
	var extraDocument any
	if err := decoder.Decode(&extraDocument); err != io.EOF {
		if err != nil {
			return Pipeline{}, fmt.Errorf("decode YAML: %w", err)
		}
		return Pipeline{}, fmt.Errorf("configuration must contain exactly one YAML document")
	}

	return pipeline, nil
}
