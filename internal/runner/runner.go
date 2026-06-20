// Package runner executes Entaxi pipelines and reports their results.
package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/jorgeSia/entaxi-ci/internal/pipeline"
)

// Runner executes pipelines and streams their command output.
type Runner struct {
	// Out receives normal runner and command output.
	Out io.Writer
	// Err receives command error output.
	Err io.Writer
}

// New creates a runner that writes to the supplied output streams.
func New(out, err io.Writer) *Runner {
	return &Runner{
		Out: out,
		Err: err,
	}
}

// Result summarizes one completed pipeline execution.
type Result struct {
	// Passed is true only when every step exits successfully.
	Passed bool
	// Duration is the total wall-clock time spent running the pipeline.
	Duration time.Duration
	// ExitCode is zero on success or the failed command's exit code.
	ExitCode int
}

// Run executes a pipeline's steps sequentially and stops at the first failure.
func (r *Runner) Run(ctx context.Context, p pipeline.Pipeline) (Result, error) {
	// Measure the whole pipeline separately from each individual step.
	started := time.Now()

	fmt.Fprintf(r.Out, "Running pipeline in %s\n\n", p.Dir)

	// Pipeline order matters, so steps run one at a time in declaration order.
	for i, step := range p.Steps {
		stepStarted := time.Now()

		fmt.Fprintf(r.Out, "[%d/%d] %s\n", i+1, len(p.Steps), step.Name)
		fmt.Fprintf(r.Out, "$ %s\n", step.Command)

		// Stream stdout and stderr while the command runs instead of buffering them.
		exitCode, err := r.runCommand(ctx, p.Dir, step.Command)
		duration := time.Since(stepStarted)
		if err != nil {
			// A real exit code means the user's command ran and the build failed.
			if exitCode >= 0 {
				total := time.Since(started)
				fmt.Fprintf(r.Out, "\nStep failed in %s with exit code %d\n", formatDuration(duration), exitCode)
				fmt.Fprintf(r.Out, "Build failed in %s\n", formatDuration(total))
				return Result{Passed: false, Duration: total, ExitCode: exitCode}, nil
			}

			// No exit code means Entaxi could not execute the command correctly.
			return Result{Passed: false, Duration: time.Since(started), ExitCode: exitCode}, err
		}

		fmt.Fprintf(r.Out, "\nStep passed in %s\n\n", formatDuration(duration))
	}

	// Reaching the end means every configured step completed successfully.
	total := time.Since(started)
	fmt.Fprintf(r.Out, "Build passed in %s\n", formatDuration(total))

	return Result{Passed: true, Duration: total, ExitCode: 0}, nil
}

// runCommand executes one shell command in the pipeline's working directory.
func (r *Runner) runCommand(ctx context.Context, dir, command string) (int, error) {
	// CommandContext allows cancellation, while sh -c supports normal shell syntax.
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = r.Out
	cmd.Stderr = r.Err

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}

	// ExitError proves that the process started and returned a non-zero status.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), err
	}

	// Use -1 when no process exit code exists, which identifies a system error.
	return -1, err
}

// formatDuration keeps human-facing timings short and stable.
func formatDuration(duration time.Duration) string {
	return duration.Round(100 * time.Millisecond).String()
}
