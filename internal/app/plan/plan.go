// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package plan contains the app definition for act-tester-plan.
package plan

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/MattWindsor91/act-tester/internal/model/id"

	"github.com/1set/gut/ystring"
	"github.com/MattWindsor91/act-tester/internal/model/machine"

	"github.com/1set/gut/yos"

	"github.com/MattWindsor91/act-tester/internal/act"

	"github.com/MattWindsor91/act-tester/internal/config"
	"github.com/MattWindsor91/act-tester/internal/plan"
	"github.com/MattWindsor91/act-tester/internal/stage/planner"
	"github.com/MattWindsor91/act-tester/internal/ux/singleobs"
	"github.com/MattWindsor91/act-tester/internal/ux/stdflag"
	c "github.com/urfave/cli/v2"
)

const (
	flagCompilerFilter  = "filter-compilers"
	usageCompilerFilter = "`glob` to use to filter compilers to enable"

	flagMachineFilter  = "filter-machines"
	usageMachineFilter = "`glob` to use to filter machines to plan"
)

// App creates the act-tester-plan app.
func App(outw, errw io.Writer) *c.App {
	a := c.App{
		Name:  "act-tester-plan",
		Usage: "runs the planning phase of an ACT test standalone",
		Flags: flags(),
		Action: func(ctx *c.Context) error {
			return run(ctx, os.Stdout, os.Stderr)
		},
	}
	return stdflag.SetCommonAppSettings(&a, outw, errw)
}

func flags() []c.Flag {
	ownFlags := []c.Flag{
		stdflag.ConfFileCliFlag(),
		&c.StringFlag{
			Name:        flagCompilerFilter,
			Aliases:     []string{stdflag.FlagCompiler},
			Usage:       usageCompilerFilter,
			DefaultText: "all compilers",
		},
		&c.StringFlag{
			Name:        flagMachineFilter,
			Aliases:     []string{stdflag.FlagMachine},
			Usage:       usageMachineFilter,
			DefaultText: "all machines",
		},
		stdflag.WorkerCountCliFlag(),
		stdflag.OutDirCliFlag(""),
	}
	return append(ownFlags, stdflag.ActRunnerCliFlags()...)
}

func run(ctx *c.Context, outw, errw io.Writer) error {
	cfg, err := stdflag.ConfFileFromCli(ctx)
	if err != nil {
		return err
	}

	pr, err := makePlanner(ctx, cfg, errw)
	if err != nil {
		return err
	}

	ms, err := machines(ctx, cfg)
	if err != nil {
		return err
	}
	dir, err := outDir(ctx, ms)
	if err != nil {
		return err
	}

	ps, err := pr.Plan(ctx.Context, ms, ctx.Args().Slice()...)
	if err != nil {
		return err
	}

	return writePlans(outw, dir, ps)
}

func machines(ctx *c.Context, cfg *config.Config) (machine.ConfigMap, error) {
	midstr := ctx.String(flagMachineFilter)
	if ystring.IsBlank(midstr) {
		return cfg.Machines, nil
	}
	return globbedMachines(midstr, cfg.Machines)
}

func globbedMachines(midstr string, configMap machine.ConfigMap) (machine.ConfigMap, error) {
	mid, err := id.TryFromString(midstr)
	if err != nil {
		return nil, err
	}
	return configMap.Filter(mid)
}

func outDir(ctx *c.Context, ms map[string]machine.Config) (string, error) {
	dir := stdflag.OutDirFromCli(ctx)
	if ystring.IsBlank(dir) && len(ms) != 1 {
		return "", fmt.Errorf("must specify directory if planning multiple machines (have %d)", len(ms))
	}
	return dir, nil
}

func writePlans(outw io.Writer, outdir string, ps map[string]plan.Plan) error {
	// Assuming that outDir above has dealt with the case whereby there is no output directory but multiple plans.
	if ystring.IsBlank(outdir) {
		return writePlansToWriter(outw, ps)
	}
	return writePlansToDir(outdir, ps)
}

func writePlansToWriter(w io.Writer, ps map[string]plan.Plan) error {
	for _, p := range ps {
		if err := p.Write(w, plan.WriteHuman); err != nil {
			return err
		}
	}
	return nil
}

func writePlansToDir(outdir string, ps map[string]plan.Plan) error {
	if err := yos.MakeDir(outdir); err != nil {
		return err
	}
	for n, p := range ps {
		file := fmt.Sprintf("plan.%s.json", n)
		if err := p.WriteFile(yos.JoinPath(outdir, file), plan.WriteHuman); err != nil {
			return err
		}
	}
	return nil
}

func makePlanner(ctx *c.Context, cfg *config.Config, errw io.Writer) (*planner.Planner, error) {
	a := stdflag.ActRunnerFromCli(ctx, errw)

	qs := quantities(ctx)
	src := source(a, cfg)

	l := log.New(errw, "[planner] ", log.LstdFlags)

	return planner.New(
		src,
		planner.ObserveWith(singleobs.Planner(l)...),
		planner.OverrideQuantities(qs),
		planner.FilterCompilers(ctx.String(flagCompilerFilter)),
	)
}

func source(a *act.Runner, cfg *config.Config) planner.Source {
	return planner.Source{
		BProbe:  cfg,
		CLister: cfg.Machines,
		SProbe:  a,
	}
}

func quantities(ctx *c.Context) planner.QuantitySet {
	return planner.QuantitySet{
		NWorkers: stdflag.WorkerCountFromCli(ctx),
	}
}
