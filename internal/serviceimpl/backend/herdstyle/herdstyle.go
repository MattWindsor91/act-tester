// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

// Package herdstyle contains backends that act in a similar way to the Herd memory simulator.
//
// Herd is a de facto standard in the area of concurrency exploration, so various tools have the same flow, which
// has the following characteristics:
//
// - Is an external, third-party tool running on the local machine, largely communicated with by command-line flags;
//
// - Accepts Litmus tests (different tools accept different architectures, possibly including C);
//
// - Outputs observations to stdout in a loosely standard format (handled by the parser package);
//
// - Generally run standalone, though some tools may support lifting to executables.
package herdstyle

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/c4-project/c4t/internal/model/id"

	"github.com/c4-project/c4t/internal/helper/errhelp"

	backend2 "github.com/c4-project/c4t/internal/model/service/backend"

	"github.com/c4-project/c4t/internal/serviceimpl/backend/herdstyle/parser"

	"github.com/c4-project/c4t/internal/model/recipe"

	"github.com/c4-project/c4t/internal/model/service"
	"github.com/c4-project/c4t/internal/subject/obs"
)

// standaloneOut is the name of the file in the output directory to which we should write standalone output.
const standaloneOut = "output.txt"

// Backend represents herd-style backends such as Herd and Litmus.
type Backend struct {
	// OptCapabilities contains the capability flags for this backend not implied by being a herdstyle backend.
	OptCapabilities backend2.Capability

	// Arches describes the architectures of Litmus test this backend can deal with.
	Arches []id.ID

	// RunInfo is the run information for the particular backend.
	RunInfo service.RunInfo

	// Impl provides parts of the Backend backend setup that differ between the various tools.
	Impl BackendImpl
}

// Instantiate overrides the run info in this backend, and returns a new backend in an interface wrapper.
//
// This slightly strange implementation is designed to slot into function tables in resolvers.
func (h Backend) Instantiate(new *service.RunInfo) backend2.Backend {
	h.RunInfo.OverrideIfNotNil(new)
	return h
}

// Capabilities returns OptCapabilities (as well as the implied backend.CanLiftLitmus and backend.CanRunStandalone).
func (h Backend) Capabilities() backend2.Capability {
	return backend2.CanLiftLitmus | backend2.CanRunStandalone | h.OptCapabilities
}

// LitmusArches returns Arches, to satisfy the backend interface.
func (h Backend) LitmusArches() []id.ID {
	return h.Arches
}

// ParseObs parses an observation from r into o.
func (h Backend) ParseObs(_ context.Context, r io.Reader, o *obs.Obs) error {
	return parser.Parse(h.Impl, r, o)
}

func (h Backend) Lift(ctx context.Context, j backend2.LiftJob, x service.Runner) (recipe.Recipe, error) {
	if err := h.checkAndAmendJob(&j); err != nil {
		return recipe.Recipe{}, err
	}
	switch j.Out.Target {
	case backend2.ToStandalone:
		return h.liftStandalone(ctx, j, x)
	case backend2.ToExeRecipe:
		return h.liftExe(ctx, j, x)
	}
	// We should've filtered out bad targets by this stage.
	return recipe.Recipe{}, nil
}

func (h Backend) liftStandalone(ctx context.Context, j backend2.LiftJob, x service.Runner) (recipe.Recipe, error) {
	if err := h.runStandalone(ctx, j, x); err != nil {
		return recipe.Recipe{}, err
	}
	return h.makeStandaloneRecipe(j.Out)
}

func (h Backend) liftExe(ctx context.Context, j backend2.LiftJob, x service.Runner) (recipe.Recipe, error) {
	if err := h.Impl.LiftExe(ctx, j, h.RunInfo, x); err != nil {
		return recipe.Recipe{}, err
	}
	return h.makeExeRecipe(j.Out)
}

func (h Backend) runStandalone(ctx context.Context, j backend2.LiftJob, x service.Runner) error {
	f, err := os.Create(filepath.Join(filepath.Clean(j.Out.Dir), standaloneOut))
	if err != nil {
		return fmt.Errorf("couldn't create standalone output file: %s", err)
	}
	rerr := h.Impl.LiftStandalone(ctx, j, h.RunInfo, x, f)
	cerr := f.Close()
	return errhelp.FirstError(rerr, cerr)
}

func (h Backend) checkAndAmendJob(j *backend2.LiftJob) error {
	if err := j.Check(); err != nil {
		return err
	}

	if !j.Arch.IsEmpty() && !j.In.Litmus.IsC() {
		return fmt.Errorf("%w: can only set lifting architecture for C litmus tests", backend2.ErrNotSupported)
	}

	if err := h.checkAndAmendInput(&j.In); err != nil {
		return err
	}
	return h.checkAndAmendOutput(&j.Out)
}

func (h Backend) checkAndAmendInput(i *backend2.LiftInput) error {
	if i.Source != backend2.LiftLitmus {
		return fmt.Errorf("%w: can only lift litmus files", backend2.ErrNotSupported)
	}
	if !h.supportsLitmusArch(i.Litmus.Arch) {
		return fmt.Errorf("%w: architecture %q not supported", backend2.ErrNotSupported, i.Litmus.Arch)
	}
	return nil
}

func (h Backend) supportsLitmusArch(a id.ID) bool {
	for _, a2 := range h.Arches {
		if a.HasPrefix(a2) {
			return true
		}
	}
	return false
}

func (h Backend) checkAndAmendOutput(o *backend2.LiftOutput) error {
	switch o.Target {
	case backend2.ToDefault:
		o.Target = backend2.ToStandalone
		fallthrough
	case backend2.ToStandalone:
	case backend2.ToExeRecipe:
		if (h.OptCapabilities & backend2.CanProduceExe) == 0 {
			return fmt.Errorf("%w: cannot produce executables", backend2.ErrNotSupported)
		}
	case backend2.ToObjRecipe:
		return fmt.Errorf("%w: cannot produce objects", backend2.ErrNotSupported)
	}
	return nil
}

func (h Backend) makeStandaloneRecipe(o backend2.LiftOutput) (recipe.Recipe, error) {
	return recipe.New(o.Dir,
		recipe.OutNothing,
		recipe.AddFiles(standaloneOut),
	)
}

func (h Backend) makeExeRecipe(o backend2.LiftOutput) (recipe.Recipe, error) {
	fs, err := o.Files()
	if err != nil {
		return recipe.Recipe{}, err
	}

	return recipe.New(o.Dir,
		recipe.OutExe,
		recipe.AddFiles(fs...),
		// TODO(@MattWindsor91): delitmus support
		recipe.CompileAllCToExe(),
	)
}

// BackendImpl describes the functionality that differs between Herdtools-style backends.
type BackendImpl interface {
	// LiftStandalone runs the lifter job j using x and the run information in r, expecting it to output the
	// results into w.
	LiftStandalone(ctx context.Context, j backend2.LiftJob, r service.RunInfo, x service.Runner, w io.Writer) error

	// LiftExe runs the lifter job j using x and the run information in r, expecting an executable.
	LiftExe(ctx context.Context, j backend2.LiftJob, r service.RunInfo, x service.Runner) error

	parser.Impl
}
