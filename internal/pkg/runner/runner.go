// Package runner contains the logic for the single-file test runner.
package runner

import (
	"context"

	"github.com/MattWindsor91/act-tester/internal/pkg/model"
)

// Compiler is the interface of things that can run compilers.
type Compiler interface {
	// Run runs the compiler pointed to by compiler on the input files infiles, outputting a binary to outfile.
	Compile(compiler model.ID, infiles string, outfile string) error
}

// Runner contains the configuration required to perform a single test run.
type Runner struct {
	// Plan is the machine plan on which this runner is operating.
	Plan model.MachinePlan

	// Compiler is the compiler runner that we're using to do this test run.
	Compiler Compiler

	// OutDir contains the root output directory for things generated by this runner.
	OutDir string
}

// RunPlanFile loads a single-machine plan from path into this Runner and runs it.
func (r *Runner) RunPlanFile(ctx context.Context, path string) error {
	// TODO(@MattWindsor91): load plan
	return r.Run(ctx)
}

// Run runs the runner, assuming that one has loaded a valid machine plan.
func (r *Runner) Run(ctx context.Context) error {
	return nil
}
