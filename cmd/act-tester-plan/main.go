// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/MattWindsor91/act-tester/internal/pkg/model"

	"github.com/MattWindsor91/act-tester/internal/pkg/act"
	"github.com/MattWindsor91/act-tester/internal/pkg/ux"

	"github.com/MattWindsor91/act-tester/internal/pkg/planner"
)

const (
	usageCompPred = "predicate `sexp` used to filter compilers for this test plan"
	usageMach     = "ID of machine to use for this test plan"
)

func main() {
	err := run(os.Args, os.Stdout, os.Stderr)
	ux.LogTopError(err)
}

func run(args []string, outw, errw io.Writer) error {
	a := act.Runner{Stderr: errw}
	l := log.New(errw, "", 0)
	plan := planner.Planner{
		Source:   &a,
		Logger:   l,
		Observer: ux.NewPbObserver(l),
	}

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.StringVar(&plan.Filter, "c", "", usageCompPred)
	pmach := fs.String(ux.FlagMachine, "", usageMach)
	ux.ActRunnerFlags(fs, &a)
	ux.CorpusSizeFlag(fs, &plan.CorpusSize)

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	plan.InFiles = fs.Args()
	plan.MachineID = model.IDFromString(*pmach)

	p, err := plan.Plan(context.Background())
	if err != nil {
		return err
	}
	return p.Dump(outw)
}
