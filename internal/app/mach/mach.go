// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package mach contains the app definition for act-tester-mach.
package mach

import (
	"encoding/json"
	"io"
	"log"
	"strings"

	"github.com/MattWindsor91/act-tester/internal/app/rmach"
	"github.com/MattWindsor91/act-tester/internal/controller/mach"
	"github.com/MattWindsor91/act-tester/internal/controller/mach/forward"
	"github.com/MattWindsor91/act-tester/internal/helper/iohelp"
	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"
	bimpl "github.com/MattWindsor91/act-tester/internal/serviceimpl/backend"
	cimpl "github.com/MattWindsor91/act-tester/internal/serviceimpl/compiler"
	"github.com/MattWindsor91/act-tester/internal/view"
	"github.com/MattWindsor91/act-tester/internal/view/singleobs"
	"github.com/MattWindsor91/act-tester/internal/view/stdflag"
	c "github.com/urfave/cli/v2"
)

const (
	Name = "act-tester-mach"

	readme = `
   This part of the tester, also known as the 'machine invoker', runs the parts
   of a testing cycle that are specific to the machine-under-test.

   This command's target audience is a pipe, possibly over SSH, connected to an
   instance of the ` + rmach.Name + ` command.  As such, it doesn't make many
   efforts to be user-friendly, and you probably want to use that command
   instead.
`
)

// App creates the act-tester-mach app.
func App(outw, errw io.Writer) *c.App {
	a := c.App{
		Name:        Name,
		Usage:       "runs the machine-dependent phase of an ACT test",
		Description: strings.TrimSpace(readme),
		Flags:       flags(),
		Action: func(ctx *c.Context) error {
			return run(ctx, outw, errw)
		},
	}
	return stdflag.SetPlanAppSettings(&a, outw, errw)
}

func flags() []c.Flag {
	return stdflag.MachCliFlags()
}

func run(ctx *c.Context, outw, errw io.Writer) error {
	cfg := makeConfig(ctx, outw, errw)
	return view.RunOnCliPlan(ctx, cfg, outw)
}

func makeConfig(ctx *c.Context, outw, errw io.Writer) *mach.Config {
	cfg := mach.Config{
		CDriver: &cimpl.CResolve,
		RDriver: &bimpl.BResolve,
		Stdout:  outw,
		User:    stdflag.MachConfigFromCli(ctx, mach.QuantitySet{}),
	}
	setLoggerAndObservers(&cfg, errw, ctx.Bool(stdflag.FlagUseJSONLong))
	return &cfg
}

func setLoggerAndObservers(c *mach.Config, errw io.Writer, jsonStatus bool) {
	errw = iohelp.EnsureWriter(errw)

	if jsonStatus {
		c.Logger = nil
		c.Observers = makeJsonObserver(errw)
		return
	}

	c.Logger = log.New(errw, "[mach] ", log.LstdFlags)
	c.Observers = singleobs.Builder(c.Logger)
}

func makeJsonObserver(errw io.Writer) []builder.Observer {
	return []builder.Observer{&forward.Observer{Encoder: json.NewEncoder(errw)}}
}