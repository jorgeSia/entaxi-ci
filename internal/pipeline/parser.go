package pipeline

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// parse decodes YAML data into a pipeline without performing filesystem work.
func parse(data []byte) (Pipeline, error) {
	var pipeline Pipeline
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return Pipeline{}, fmt.Errorf("parse %s: %w", FileName, err)
	}

	return pipeline, nil
}
