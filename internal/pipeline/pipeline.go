// Package pipeline defines and loads Entaxi pipeline configuration.
package pipeline

import (
	"fmt"
	"strings"

	"github.com/jorgeSia/entaxi-ci/internal/filemanager"
)

// FileName is the pipeline configuration file Entaxi looks for in a project.
const FileName = ".entaxi.yaml"

// Pipeline describes the commands to run and the directory in which to run them.
type Pipeline struct {
	// Name is the optional human-readable pipeline name from .entaxi.yaml.
	Name string `yaml:"name"`
	// Dir is resolved at load time and is not read from YAML.
	Dir string `yaml:"-"`
	// Steps are executed sequentially in their declared order.
	Steps []Step `yaml:"steps"`
}

// Step describes one shell command in a pipeline.
type Step struct {
	// Name identifies the step in CLI output.
	Name string `yaml:"name"`
	// Command is passed to the system shell for execution.
	Command string `yaml:"command"`
}

// New creates a ready-to-run pipeline from the configuration in dir.
func New(dir string) (Pipeline, error) {
	// Resolve and validate the project directory before looking for configuration.
	absDir, err := filemanager.ResolveDirectory(dir)
	if err != nil {
		return Pipeline{}, err
	}

	// Delegate path construction and file reading to the filesystem boundary.
	configPath := filemanager.Join(absDir, FileName)
	data, err := filemanager.ReadFile(configPath)
	if err != nil {
		return Pipeline{}, err
	}

	// Delegate YAML syntax handling to the parser.
	pipeline, err := parse(data)
	if err != nil {
		return Pipeline{}, fmt.Errorf("parse pipeline config %q: %w", configPath, err)
	}

	// Add runtime context that is intentionally not sourced from YAML.
	pipeline.Dir = absDir

	// Return only pipelines that are normalized and safe for the runner to consume.
	if err := pipeline.Validate(); err != nil {
		return Pipeline{}, fmt.Errorf("validate pipeline config %q: %w", configPath, err)
	}

	return pipeline, nil
}

// Validate normalizes a pipeline and rejects configurations that cannot run.
func (p *Pipeline) Validate() error {
	// Normalize the optional display name without changing internal whitespace.
	p.Name = strings.TrimSpace(p.Name)

	// A pipeline without steps has no useful work to perform.
	if len(p.Steps) == 0 {
		return fmt.Errorf("%s must define at least one step", FileName)
	}

	// Normalize each step and check all required values in declaration order.
	for i := range p.Steps {
		step := &p.Steps[i]
		step.Name = strings.TrimSpace(step.Name)
		step.Command = strings.TrimSpace(step.Command)

		// Missing names are safe to default because names affect only presentation.
		if step.Name == "" {
			step.Name = fmt.Sprintf("Step %d", i+1)
		}
		// A missing command makes the step impossible to execute.
		if step.Command == "" {
			return fmt.Errorf("step %d (%q) must define a command", i+1, step.Name)
		}
	}

	return nil
}
