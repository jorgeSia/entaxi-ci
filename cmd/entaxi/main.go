package main

import (
	"os"

	"github.com/jorgeSia/entaxi-ci/internal/cli"
)

func main() {
	// Run the CLI with the process arguments and return its result to the shell.
	os.Exit(cli.Default().Run(os.Args[1:]))
}
