// Package cli provides the main functionalities of the command line interface.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jorgeSia/entaxi-ci/internal/pipeline"
	"github.com/jorgeSia/entaxi-ci/internal/runner"
)

const usage = `Entaxi checks whether local projects are in a good state.

Usage:
  entaxi run [path]

Commands:
  run [path]    Run the .entaxi.yaml pipeline in path, or the current directory
`

type CLI struct {
	// Out receives normal command output.
	Out io.Writer
	// Err receives errors and usage mistakes.
	Err io.Writer
}

// New creates a CLI that writes to the supplied output streams.
func New(out, err io.Writer) *CLI {
	return &CLI{
		Out: out,
		Err: err,
	}
}

// Default creates a CLI connected to the process standard output and error.
func Default() *CLI {
	return New(os.Stdout, os.Stderr)
}

// Run dispatches a command and returns the exit code that main should report.
func (c *CLI) Run(args []string) int {
	// A command is required, so missing arguments are reported as a usage error.
	if len(args) == 0 {
		fmt.Fprint(c.Err, usage)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(c.Out, usage)
		return 0
	case "run":
		return c.runPipeline(args[1:])
	default:
		fmt.Fprintf(c.Err, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func (c *CLI) runPipeline(args []string) int {
	// The run command accepts only one optional project directory.
	if len(args) > 1 {
		fmt.Fprintf(c.Err, "run expects at most one path\n\n%s", usage)
		return 2
	}

	// Run the current directory unless the user supplied another path.
	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}

	// Create and validate the project's pipeline before executing anything.
	p, err := pipeline.New(dir)
	if err != nil {
		fmt.Fprintf(c.Err, "failed to load pipeline: %v\n", err)
		return 1
	}

	// Execute every configured step and stream command output through the CLI.
	result, err := runner.Run(context.Background(), p, c.Out, c.Err)
	if err != nil {
		fmt.Fprintf(c.Err, "build error: %v\n", err)
		return 1
	}
	// A command failure is an expected build result, but still a non-zero CLI exit.
	if !result.Passed {
		return 1
	}

	return 0
}
