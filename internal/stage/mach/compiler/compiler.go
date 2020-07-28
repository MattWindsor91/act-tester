// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package compiler contains a test-plan batch compiler.
// It relies on the existence of a single-binary compiler.
package compiler

import (
	"context"
	"errors"
	"io"

	"github.com/MattWindsor91/act-tester/internal/stage/mach/quantity"

	"github.com/MattWindsor91/act-tester/internal/plan/stage"

	"github.com/MattWindsor91/act-tester/internal/model/job/compile"
	"github.com/MattWindsor91/act-tester/internal/model/subject"

	"github.com/MattWindsor91/act-tester/internal/model/service/compiler"

	"github.com/MattWindsor91/act-tester/internal/model/id"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"

	"github.com/MattWindsor91/act-tester/internal/model/corpus"

	"golang.org/x/sync/errgroup"

	"github.com/MattWindsor91/act-tester/internal/helper/iohelp"

	"github.com/MattWindsor91/act-tester/internal/plan"
)

// ErrDriverNil occurs when the compiler tries to use the nil pointer as its single-compile driver.
var ErrDriverNil = errors.New("driver nil")

// Driver is the interface of things that can run compilers.
type Driver interface {
	// RunCompiler runs the compiler job j.
	// If applicable, errw will be connected to the compiler's standard error.
	//
	// Implementors should note that the paths in j are slash-paths, and will need converting to filepaths.
	RunCompiler(ctx context.Context, j compile.Single, errw io.Writer) error
}

//go:generate mockery --name=Driver

// SubjectPather is the interface of types that can produce path sets for compilations.
type SubjectPather interface {
	// Prepare sets up the directories ready to serve through SubjectPaths.
	// It takes the list of compiler IDs that are to be represented in the pathset.
	Prepare(compilers []id.ID) error

	// SubjectPaths gets the binary and log file paths for the subject/compiler pair sc.
	SubjectPaths(sc SubjectCompile) subject.CompileFileset
}

//go:generate mockery --name=SubjectPather

// Compiler contains the configuration required to compile the recipes for a single test run.
type Compiler struct {
	// driver is what the compiler should use to run single compiler jobs.
	driver Driver

	// observers observe the compiler's progress across a corpus.
	observers []builder.Observer

	// paths is the pathset for this compiler run.
	paths SubjectPather

	// quantities is this compiler stage's quantity set.
	quantities quantity.SingleSet
}

// New creates a new batch compiler instance using the config c and plan p.
// It can fail if various safety checks fail on the config,
// or if there is no obvious machine that the compiler can target.
func New(driver Driver, paths SubjectPather, opts ...Option) (*Compiler, error) {
	if driver == nil {
		return nil, ErrDriverNil
	}
	if paths == nil {
		return nil, iohelp.ErrPathsetNil
	}

	c := &Compiler{driver: driver, paths: paths}
	if err := Options(opts...)(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Run runs the batch compiler with context ctx and plan p.
// On success, it returns an amended plan, now associating each subject with its compiler results.
func (c *Compiler) Run(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	if err := checkPlan(p); err != nil {
		return nil, err
	}
	return p.RunStage(ctx, stage.Compile, c.runInner)
}

func (c *Compiler) runInner(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	if err := c.prepareDirs(p); err != nil {
		return nil, err
	}

	eg, ectx := errgroup.WithContext(ctx)

	b, err := c.makeBuilder(p)
	if err != nil {
		return nil, err
	}

	// TODO(@MattWindsor91): port this to observers
	// c.quantities.Log(c.l)

	for ids, cc := range p.Compilers {
		nc, err := cc.AddNameString(ids)
		if err != nil {
			return nil, err
		}
		cr := c.instance(b.SendCh, nc, p)
		eg.Go(func() error {
			return cr.Compile(ectx)
		})
	}

	var newc corpus.Corpus
	eg.Go(func() error {
		var err error
		newc, err = b.Run(ectx)
		return err
	})

	// Need to wait until there are no goroutines accessing the corpus before we copy it over.
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	np := *p
	np.Corpus = newc
	return &np, nil
}

func checkPlan(p *plan.Plan) error {
	if p == nil {
		return plan.ErrNil
	}
	return p.Check()
}

func (c *Compiler) makeBuilder(p *plan.Plan) (*builder.Builder, error) {
	bc := builder.Config{
		Init:      p.Corpus,
		Observers: c.observers,
		Manifest: builder.Manifest{
			Name:  "compile",
			NReqs: p.NumExpCompilations(),
		},
	}
	return builder.New(bc)
}

func (c *Compiler) prepareDirs(p *plan.Plan) error {
	// TODO(@MattWindsor91): port this to observers
	//c.l.Println("preparing directories")
	cids, err := p.CompilerIDs()
	if err != nil {
		return err
	}
	return c.paths.Prepare(cids)
}

// instance makes an instance for the named compiler nc, outputting results to resCh.
// It also takes in a read-only copy, rc, of the corpus; this is because the result handling thread will be modifying
// the corpus proper.
func (c *Compiler) instance(requests chan<- builder.Request, nc *compiler.Named, p *plan.Plan) *Instance {
	return &Instance{
		machineID: p.Machine.ID,
		compiler:  nc,
		corpus:    p.Corpus,
		driver:    c.driver,
		paths:     c.paths,
		resCh:     requests,
	}
}
