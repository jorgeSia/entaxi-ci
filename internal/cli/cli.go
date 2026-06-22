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

const (
	exitSuccess = 0
	exitError   = 1
	exitUsage   = 2
)

const rootUsage = `Entaxi checks whether local projects are in a good state.

Usage:
  entaxi <command> [arguments]
  entaxi help [command]

Commands:
  run [path]    Run the .entaxi.yaml pipeline in path, or the current directory
  help [command] Show help for Entaxi or a command
`

const runUsage = `Run the pipeline defined in a project's .entaxi.yaml file.

Usage:
  entaxi run [path]
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
		fmt.Fprint(c.Err, rootUsage)
		return exitUsage
	}

	switch args[0] {
	case "help", "-h", "--help":
		return c.help(args[1:])
	case "run":
		return c.runPipeline(args[1:])
	default:
		fmt.Fprintf(c.Err, "unknown command %q\n\n%s", args[0], rootUsage)
		return exitUsage
	}
}

func (c *CLI) help(args []string) int {
	// Help accepts one optional command name.
	if len(args) > 1 {
		fmt.Fprintf(c.Err, "help expects at most one command\n\n%s", rootUsage)
		return exitUsage
	}

	if len(args) == 0 {
		fmt.Fprint(c.Out, rootUsage)
		return exitSuccess
	}

	// Each command owns its detailed usage text.
	switch args[0] {
	case "run":
		fmt.Fprint(c.Out, runUsage)
		return exitSuccess
	default:
		fmt.Fprintf(c.Err, "unknown help topic %q\n\n%s", args[0], rootUsage)
		return exitUsage
	}
}

func (c *CLI) runPipeline(args []string) int {
	// Treat help flags as command help rather than project paths.
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprint(c.Out, runUsage)
		return exitSuccess
	}

	// The run command accepts only one optional project directory.
	if len(args) > 1 {
		fmt.Fprintf(c.Err, "run expects at most one path\n\n%s", runUsage)
		return exitUsage
	}

	// Run the current directory unless the user supplied another path.
	dir := "."
	if len(args) == 1 {
		dir = args[0]
	}

	// Create and validate the project's pipeline before executing anything.
	p, err := pipeline.New(dir)
	if err != nil {
		fmt.Fprintf(c.Err, "failed to create pipeline: %v\n", err)
		return exitError
	}

	// Give this execution its own runner while reusing the CLI output streams.
	r := runner.New(c.Out, c.Err)
	result, err := r.Run(context.Background(), p)
	if err != nil {
		fmt.Fprintf(c.Err, "build error: %v\n", err)
		return exitError
	}
	// A command failure is an expected build result, but still a non-zero CLI exit.
	if !result.Passed {
		return exitError
	}

	return exitSuccess
}
