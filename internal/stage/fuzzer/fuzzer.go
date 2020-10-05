// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package fuzzer contains a test-plan batch fuzzer.
// It relies on the existence of a single-file fuzzer such as act-fuzz.
package fuzzer

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/MattWindsor91/act-tester/internal/machine"

	"github.com/MattWindsor91/act-tester/internal/quantity"

	"github.com/MattWindsor91/act-tester/internal/plan/stage"

	"github.com/MattWindsor91/act-tester/internal/subject/corpus/builder"

	"github.com/MattWindsor91/act-tester/internal/subject/corpus"

	"github.com/MattWindsor91/act-tester/internal/subject"

	"github.com/MattWindsor91/act-tester/internal/helper/iohelp"

	"github.com/MattWindsor91/act-tester/internal/plan"
)

// DefaultSubjectCycles is the default number of fuzz cycles to run per subject.
const DefaultSubjectCycles = 10

// SubjectPather is the interface of things that serve file-paths for subject outputs during a fuzz batch.
type SubjectPather interface {
	// Prepare sets up the directories ready to serve through SubjectPaths.
	Prepare() error
	// SubjectLitmus gets the litmus filepath for the subject/cycle pair sc.
	SubjectLitmus(sc SubjectCycle) string
	// SubjectTrace gets the trace filepath for the subject/cycle pair sc.
	SubjectTrace(sc SubjectCycle) string
}

//go:generate mockery --name=SubjectPather

// Fuzzer holds the state required for the fuzzing stage of the tester.
type Fuzzer struct {
	// l is the logger for this fuzzer.
	l *log.Logger
	// driver holds the fuzzer's low-level implementation structs.
	driver Driver
	// observers observe the fuzzer's progress across a corpus.
	observers []builder.Observer
	// paths contains the path set for things generated by this fuzzer.
	paths SubjectPather
	// quantities sets the quantities for this batch fuzzer run.
	quantities quantity.FuzzSet
}

// New constructs a fuzzer with the config c and plan p.
func New(d Driver, ps SubjectPather, o ...Option) (*Fuzzer, error) {
	f := Fuzzer{
		driver: d,
		paths:  ps,
		quantities: quantity.FuzzSet{
			SubjectCycles: DefaultSubjectCycles,
		},
	}
	if err := Options(o...)(&f); err != nil {
		return nil, err
	}
	f.l = iohelp.EnsureLog(f.l)
	return &f, f.check()
}

// check makes sure various parts of this fuzzer's config are present.
func (f *Fuzzer) check() error {
	if f.driver == nil {
		return ErrDriverNil
	}
	if f.paths == nil {
		return iohelp.ErrPathsetNil
	}
	if f.quantities.SubjectCycles <= 0 {
		return fmt.Errorf("%w: non-positive subject cycle amount", corpus.ErrSmall)
	}
	return nil
}

// Run runs the fuzzer with context ctx and plan p.
func (f *Fuzzer) Run(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	if err := f.checkPlan(p); err != nil {
		return nil, err
	}
	return p.RunStage(ctx, stage.Fuzz, f.fuzz)
}

func (f *Fuzzer) fuzz(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	f.l.Println("preparing directories")
	if err := f.paths.Prepare(); err != nil {
		return nil, err
	}

	f.l.Println("now fuzzing")
	rng := p.Metadata.Rand()
	fcs, ferr := f.fuzzCorpus(ctx, rng, p.Corpus, p.Machine.Machine)
	if ferr != nil {
		return nil, ferr
	}

	return f.sampleAndUpdatePlan(fcs, rng, *p)
}

// sampleAndUpdatePlan samples fcs, updates the fresh plan copy p with it, and returns a pointer to it.
func (f *Fuzzer) sampleAndUpdatePlan(fcs corpus.Corpus, rng *rand.Rand, p plan.Plan) (*plan.Plan, error) {
	f.l.Println("sampling corpus")
	scs, err := fcs.Sample(rng, f.quantities.CorpusSize)
	if err != nil {
		return nil, err
	}

	f.l.Println("updating plan")
	p.Corpus = scs
	// Previously, we reset the plan creation date and seed here.  This seems a little arbitrary in hindsight,
	// so we no longer do so.
	return &p, nil
}

// count counts the number of subjects and individual fuzz runs to expect from this fuzzer.
func (f *Fuzzer) count(c corpus.Corpus) (nsubjects, nruns int) {
	nsubjects = len(c)
	nruns = f.quantities.SubjectCycles * nsubjects
	return nsubjects, nruns
}

// fuzzCorpus actually does the business of fuzzing.
func (f *Fuzzer) fuzzCorpus(ctx context.Context, rng *rand.Rand, c corpus.Corpus, m machine.Machine) (corpus.Corpus, error) {
	_, nfuzzes := f.count(c)

	f.l.Printf("Fuzzing %d inputs\n", len(c))

	seeds := corpusSeeds(rng, c)

	mf := builder.Manifest{Name: "fuzz", NReqs: nfuzzes}
	bc := builder.Config{Manifest: mf, Observers: f.observers}
	return builder.ParBuild(ctx, f.quantities.NWorkers, c, bc, func(ctx context.Context, s subject.Named, ch chan<- builder.Request) error {
		return f.makeInstance(s, seeds[s.Name], m, ch).Fuzz(ctx)
	})
}

// corpusSeeds generates a seed for each subject in c using rng.
func corpusSeeds(rng *rand.Rand, c corpus.Corpus) map[string]int64 {
	seeds := make(map[string]int64)
	for n := range c {
		seeds[n] = rng.Int63()
	}
	return seeds
}

func (f *Fuzzer) makeInstance(s subject.Named, seed int64, m machine.Machine, resCh chan<- builder.Request) *Instance {
	return &Instance{
		Driver:        f.driver,
		Subject:       s,
		SubjectCycles: f.quantities.SubjectCycles,
		Pathset:       f.paths,
		Rng:           rand.New(rand.NewSource(seed)),
		ResCh:         resCh,
		Machine:       &m,
	}
}
func (f *Fuzzer) checkPlan(p *plan.Plan) error {
	if p == nil {
		return plan.ErrNil
	}
	if err := p.Check(); err != nil {
		return err
	}
	if err := p.Metadata.RequireStage(stage.Plan); err != nil {
		return err
	}
	return f.checkCount(p.Corpus)
}

func (f *Fuzzer) checkCount(c corpus.Corpus) error {
	nsubjects, nruns := f.count(c)
	if nsubjects <= 0 {
		return corpus.ErrNone
	}

	// Note that this inequality 'does the right thing' when corpus size = 0, ie no corpus size requirement.
	csize := f.quantities.CorpusSize
	if nruns < csize {
		return fmt.Errorf("%w: projected corpus size %d, want %d", corpus.ErrSmall, nruns, csize)
	}

	return nil
}
