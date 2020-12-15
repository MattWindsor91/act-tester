// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/MattWindsor91/c4t/internal/helper/errhelp"

	"github.com/MattWindsor91/c4t/internal/model/recipe"

	"github.com/MattWindsor91/c4t/internal/model/service"

	"github.com/MattWindsor91/c4t/internal/helper/srvrun"
	"github.com/MattWindsor91/c4t/internal/subject/obs"

	"github.com/MattWindsor91/c4t/internal/config"
	"github.com/MattWindsor91/c4t/internal/model/id"
	"github.com/MattWindsor91/c4t/internal/model/service/backend"
	backend2 "github.com/MattWindsor91/c4t/internal/serviceimpl/backend"
	"github.com/MattWindsor91/c4t/internal/serviceimpl/backend/resolver"
	"github.com/MattWindsor91/c4t/internal/ux/stdflag"
	c "github.com/urfave/cli/v2"
)

const (
	// Name is the name of the backend binary.
	Name  = "c4t-backend"
	usage = "runs backends standalone"

	readme = `
    This program runs lifting backends directly, in their standalone mode, and parses their results into the usual C4
    observation format.

    Doing so avoids the need to produce a testing plan, but gives less control over any compilation or other
    intermediate actions.

    The backend to run may be controlled by the -` + flagBackendIDGlob + ` glob ID, which filters on the user-defined
    name of the backend, and the -` + flagBackendStyleGlob + ` glob ID, which filters on the style of the backend.  The
    first configured backend satisfying all of the given constraints is used.`

	flagBackendIDGlob         = "backend-id"
	flagBackendIDGlobShort    = "n"
	usageBackendIDGlob        = "filter to backends whose names match `GLOB`"
	flagBackendStyleGlob      = "backend-style"
	flagBackendStyleGlobShort = "s"
	usageBackendStyleGlob     = "filter to backends whose styles match `GLOB`"
	flagArchID                = "arch"
	flagArchIDShort           = "a"
	usageArchID               = "ID of `ARCH` to target for architecture-dependent backends"
	flagDryRun                = "dry-run"
	flagDryRunShort           = "d"
	usageDryRun               = "if true, print any external commands run instead of running them"
)

// App is the c4-backend app.
func App(outw, errw io.Writer) *c.App {
	a := &c.App{
		Name:        Name,
		Usage:       usage,
		Description: readme,
		Flags:       flags(),
		Action: func(ctx *c.Context) error {
			return run(ctx, outw, errw)
		},
	}
	return stdflag.SetPlanAppSettings(a, outw, errw)
}

func flags() []c.Flag {
	ownFlags := []c.Flag{
		stdflag.ConfFileCliFlag(),
		&c.GenericFlag{Name: flagArchID, Aliases: []string{flagArchIDShort}, Usage: usageArchID, Value: &id.ID{}},
		&c.GenericFlag{Name: flagBackendIDGlob, Aliases: []string{flagBackendIDGlobShort}, Usage: usageBackendIDGlob, Value: &id.ID{}},
		&c.GenericFlag{Name: flagBackendStyleGlob, Aliases: []string{flagBackendStyleGlobShort}, Usage: usageBackendStyleGlob, Value: &id.ID{}},
		&c.BoolFlag{Name: flagDryRun, Aliases: []string{flagDryRunShort}, Usage: usageDryRun},
	}
	return append(ownFlags, stdflag.ActRunnerCliFlags()...)
}

func run(ctx *c.Context, outw io.Writer, errw io.Writer) error {
	cfg, err := stdflag.ConfFileFromCli(ctx)
	if err != nil {
		return fmt.Errorf("while getting config: %w", err)
	}
	c4f := stdflag.ActRunnerFromCli(ctx, errw)
	cri := criteriaFromCli(ctx)
	fn, err := inputNameFromCli(ctx)
	if err != nil {
		return err
	}
	arch := idFromCli(ctx, flagArchID)

	bspec, b, err := getBackend(cfg, cri)
	if err != nil {
		return err
	}

	in, err := backend.InputFromFile(ctx.Context, fn, c4f)
	if err != nil {
		return err
	}

	td, err := ioutil.TempDir("", "c4t-backend")
	if err != nil {
		return err
	}
	j := backend.LiftJob{
		Backend: bspec,
		Arch:    arch,
		In:      in,
		Out:     backend.LiftOutput{Dir: td, Target: backend.ToStandalone},
	}
	xr := makeRunner(ctx, errw)
	perr := runParseAndDump(ctx, outw, b, j, xr)
	derr := os.RemoveAll(td)
	return errhelp.FirstError(perr, derr)
}

func makeRunner(ctx *c.Context, errw io.Writer) service.Runner {
	// TODO(@MattWindsor91): the backend logic isn't very resilient against having external commands not run.
	if ctx.Bool(flagDryRun) {
		return srvrun.DryRunner{Writer: errw}
	}
	return srvrun.NewExecRunner(srvrun.StderrTo(errw))
}

func runParseAndDump(ctx *c.Context, outw io.Writer, b backend2.Backend, j backend.LiftJob, xr service.Runner) error {
	var o obs.Obs
	if err := runAndParse(ctx.Context, b, j, &o, xr); err != nil {
		return err
	}

	e := json.NewEncoder(outw)
	e.SetIndent("", "\t")
	return e.Encode(o)
}

func runAndParse(ctx context.Context, b backend2.Backend, j backend.LiftJob, o *obs.Obs, xr service.Runner) error {
	// TODO(@MattWindsor91): deduplicate with runAndParseBin?.
	r, err := b.Lift(ctx, j, xr)
	if err != nil {
		return err
	}

	if r.Output != recipe.OutNothing {
		return fmt.Errorf("can't handle recipes with outputs: %s", r.Output)
	}

	for _, fname := range r.Paths() {
		if err := parseFile(ctx, b, j, o, fname); err != nil {
			return err
		}
	}
	return nil
}

func parseFile(ctx context.Context, b backend2.Backend, j backend.LiftJob, o *obs.Obs, fname string) error {
	f, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("can't open output file %s: %w", fname, err)
	}
	perr := b.ParseObs(ctx, j.Backend, f, o)
	cerr := f.Close()
	return errhelp.FirstError(perr, cerr)
}

func inputNameFromCli(ctx *c.Context) (string, error) {
	if ctx.Args().Len() != 1 {
		return "", errors.New("expected one argument")
	}
	return ctx.Args().First(), nil
}

func getBackend(cfg *config.Config, c backend.Criteria) (*backend.Spec, backend2.Backend, error) {
	spec, err := cfg.FindBackend(c)
	if err != nil {
		return nil, nil, fmt.Errorf("while finding backend: %w", err)
	}

	s := &spec.Spec
	b, err := resolver.Resolve.Get(s)
	if err != nil {
		return nil, nil, fmt.Errorf("while resolving backend %s: %w", spec.ID, err)
	}
	return s, b, nil
}

func criteriaFromCli(ctx *c.Context) backend.Criteria {
	return backend.Criteria{
		IDGlob:    idFromCli(ctx, flagBackendIDGlob),
		StyleGlob: idFromCli(ctx, flagBackendStyleGlob),
	}
}

func idFromCli(ctx *c.Context, flag string) id.ID {
	return *(ctx.Generic(flag).(*id.ID))
}