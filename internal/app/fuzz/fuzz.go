// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package fuzz contains the app definition for act-tester-fuzz.
package fuzz

import (
	"io"
	"log"

	"github.com/MattWindsor91/act-tester/internal/controller/fuzzer"

	"github.com/MattWindsor91/act-tester/internal/view/singleobs"

	"github.com/MattWindsor91/act-tester/internal/view/stdflag"

	c "github.com/urfave/cli/v2"

	"github.com/MattWindsor91/act-tester/internal/view"
)

// defaultOutDir is the default directory used for the results of the fuzzer.
const defaultOutDir = "fuzz_results"

// App creates the act-tester-fuzz app.
func App(outw, errw io.Writer) *c.App {
	a := &c.App{
		Name:  "act-tester-fuzz",
		Usage: "runs the batch-fuzzer phase of an ACT test",
		Flags: flags(),
		Action: func(ctx *c.Context) error {
			return run(ctx, outw, errw)
		},
	}
	return stdflag.SetPlanAppSettings(a, outw, errw)
}

func flags() []c.Flag {
	fs := []c.Flag{
		stdflag.OutDirCliFlag(defaultOutDir),
		stdflag.CorpusSizeCliFlag(),
		stdflag.SubjectCyclesCliFlag(),
	}
	return append(fs, stdflag.ActRunnerCliFlags()...)
}

func run(ctx *c.Context, outw, errw io.Writer) error {
	a := stdflag.ActRunnerFromCli(ctx, errw)
	l := log.New(errw, "", 0)
	f, err := makeFuzzer(ctx, a, l)
	if err != nil {
		return err
	}
	return view.RunOnCliPlan(ctx, f, outw)
}

func makeFuzzer(ctx *c.Context, drv fuzzer.Driver, l *log.Logger) (*fuzzer.Fuzzer, error) {
	return fuzzer.New(
		drv,
		fuzzer.NewPathset(stdflag.OutDirFromCli(ctx)),
		fuzzer.LogWith(l),
		fuzzer.ObserveWith(singleobs.Builder(l)...),
		fuzzer.OverrideQuantities(setupQuantityFlags(ctx)),
	)
}

func setupQuantityFlags(ctx *c.Context) fuzzer.QuantitySet {
	return fuzzer.QuantitySet{
		CorpusSize:    stdflag.CorpusSizeFromCli(ctx),
		SubjectCycles: stdflag.SubjectCyclesFromCli(ctx),
		NWorkers:      stdflag.WorkerCountFromCli(ctx),
	}
}
