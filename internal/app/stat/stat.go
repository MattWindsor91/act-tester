// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package stat

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/c4-project/c4t/internal/stat/pretty"

	"github.com/1set/gut/ystring"
	"github.com/c4-project/c4t/internal/stat"

	"github.com/c4-project/c4t/internal/ux/stdflag"
	c "github.com/urfave/cli/v2"
)

const (
	// Name is the name of the analyser binary.
	Name  = "c4t-stat"
	usage = "inspects the statistics file"

	readme = `
   This program reads the statistics file maintained by the director, and
   prints CSV or human-readable summaries of its contents.`

	flagCsvMutations   = "csv-mutations"
	usageCsvMutations  = "dump CSV of mutation testing results"
	flagShowMutations  = "mutations"
	usageShowMutations = "show mutations matching `filter` ('all', 'hit', 'killed', 'escaped')"
	flagUseTotals      = "use-totals"
	flagUseTotalsShort = "t"
	usageUseTotals     = "use multi-session totals rather than per-session totals"
	flagStatFile       = "input"
	flagStatFileShort  = "i"
	usageStatFile      = "read statistics from this `FILE`"
)

// App is the entry point for c4t-analyse.
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
	return []c.Flag{
		stdflag.ConfFileCliFlag(),
		&c.BoolFlag{Name: flagCsvMutations, Usage: usageCsvMutations},
		&c.StringFlag{Name: flagShowMutations, Usage: usageShowMutations, DefaultText: "do not show"},
		&c.BoolFlag{Name: flagUseTotals, Aliases: []string{flagUseTotalsShort}, Usage: usageUseTotals},
		&c.PathFlag{
			Name:        flagStatFile,
			Aliases:     []string{flagStatFileShort},
			Usage:       usageStatFile,
			DefaultText: "read from configuration",
		},
	}
}

func run(ctx *c.Context, outw io.Writer, _ io.Writer) error {
	// TODO(@MattWindsor91): maybe use stat persister?
	set, err := getStats(ctx)
	if err != nil {
		return err
	}
	return dump(ctx, set, outw)
}

func dump(ctx *c.Context, set *stat.Set, w io.Writer) error {
	totals := ctx.Bool(flagUseTotals)
	csvMutations := ctx.Bool(flagCsvMutations)

	if csvMutations {
		if err := dumpCsvMutations(w, set, totals); err != nil {
			return err
		}
	}

	return prettyPrint(ctx, set, w, totals)
}

func prettyPrint(ctx *c.Context, set *stat.Set, w io.Writer, totals bool) error {
	pp, err := makePretty(ctx, w, totals)
	if pp == nil || err != nil {
		return err
	}
	return pp.Write(*set)
}

func makePretty(ctx *c.Context, w io.Writer, totals bool) (*pretty.Printer, error) {
	flt, err := makeMutationFilter(ctx)
	if flt == nil || err != nil {
		return nil, err
	}
	// TODO(@MattWindsor91): add other pretty-printing reports if needs be
	return pretty.NewPrinter(
		pretty.UseTotals(totals),
		pretty.WriteTo(w),
		pretty.ShowMutants(flt),
	)
}

func makeMutationFilter(ctx *c.Context) (stat.MutantFilter, error) {
	s := ctx.String(flagShowMutations)
	switch s {
	case "all":
		return stat.FilterAllMutants, nil
	case "hit":
		return stat.FilterHitMutants, nil
	case "killed":
		return stat.FilterKilledMutants, nil
	case "escaped":
		return stat.FilterEscapedMutants, nil
	default:
		return nil, fmt.Errorf("unsupported mutant flag: %s", s)
	}
}

func dumpCsvMutations(w io.Writer, set *stat.Set, totals bool) error {
	cw := csv.NewWriter(w)
	if err := set.DumpMutationCSVHeader(cw); err != nil {
		return err
	}
	return set.DumpMutationCSV(cw, totals)
}

func getStats(ctx *c.Context) (*stat.Set, error) {
	fname, err := getStatPath(ctx)
	if err != nil {
		return nil, err
	}

	var set stat.Set
	return &set, set.LoadFile(fname)
}

// getStatPath computes the intended path to the stats file.
func getStatPath(ctx *c.Context) (string, error) {
	if f := ctx.Path(flagStatFile); ystring.IsNotBlank(f) {
		return f, nil
	}
	cfg, err := stdflag.ConfigFromCli(ctx)
	if err != nil {
		return "", err
	}
	return cfg.Paths.StatFile()
}
