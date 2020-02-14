// Package director contains a test-plan fuzzer.
// It relies on the existence of a single-file fuzzer such as act-fuzz.
package fuzzer

import (
	"context"
	"math/rand"

	"golang.org/x/sync/errgroup"

	"github.com/MattWindsor91/act-tester/internal/pkg/plan"

	"github.com/MattWindsor91/act-tester/internal/pkg/model"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultSubjectCycles is the default number of fuzz cycles to run per subject.
	DefaultSubjectCycles = 10

	// NoChunkLimit is the chunk count that should be passed to turn off chunk limiting.
	NoChunkLimit = 0
)

// SingleFuzzer represents types that can commune with a C litmus test fuzzer.
type SingleFuzzer interface {
	// FuzzSingle fuzzes the test at path inPath using the given seed,
	// outputting the new test to path outPath and the trace to tracePath.
	FuzzSingle(seed int32, inPath, outPath, tracePath string) error
}

// Fuzzer holds the configuration required to fuzz a plan file.
type Fuzzer struct {
	// Plan is the plan on which this fuzzer is operating.
	Plan plan.Plan

	// Driver holds the single-file fuzzer that the fuzzer is going to use.
	Driver SingleFuzzer

	// Paths contains the path set for things generated by this fuzzer.
	Paths *Pathset

	// CorpusSize is the sampling size for the corpus after fuzzing.
	// It has a similar effect to CorpusSize in planner.Planner.
	CorpusSize int

	// SubjectCycles is the number of times to fuzz each file.
	SubjectCycles int

	// FuzzWorkers is the number of separate goroutines to launch for fuzzing.
	FuzzWorkers int
}

// Run runs the fuzzer, sampling the results if needed.
// Run is not thread-safe.
func (f *Fuzzer) Run(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	if err := f.prepare(p); err != nil {
		return nil, err
	}

	logrus.Infoln("now fuzzing")
	rng := rand.New(rand.NewSource(f.Plan.Seed))
	fcs, ferr := f.fuzz(ctx, rng)
	if ferr != nil {
		return nil, ferr
	}

	return f.sampleAndUpdatePlan(fcs, rng)
}

// sampleAndUpdatePlan samples fcs and places the result in the fuzzer's plan.
func (f *Fuzzer) sampleAndUpdatePlan(fcs model.Corpus, rng *rand.Rand) (*plan.Plan, error) {
	logrus.Infoln("sampling corpus")
	scs, err := fcs.Sample(rng.Int63(), f.CorpusSize)
	if err != nil {
		return nil, err
	}

	logrus.Infoln("updating plan")
	f.Plan.Corpus = scs
	f.Plan.Seed = rng.Int63()
	return &f.Plan, nil
}

// count counts the number of subjects and individual fuzz runs to expect from this fuzzer.
func (f *Fuzzer) count() (nsubjects, nruns int) {
	nsubjects = len(f.Plan.Corpus)
	nruns = f.SubjectCycles * nsubjects
	return nsubjects, nruns
}

// fuzz actually does the business of fuzzing.
func (f *Fuzzer) fuzz(ctx context.Context, rng *rand.Rand) (model.Corpus, error) {
	_, nfuzzes := f.count()

	fcs := make(model.Corpus, nfuzzes)

	eg, ectx := errgroup.WithContext(ctx)
	resCh := make(chan model.Subject)

	chunks := f.corpusChunks()
	logrus.Infof("Fuzzing %d inputs with %d chunks\n", len(f.Plan.Corpus), len(chunks))

	for _, c := range chunks {
		cp := c
		subrng := rand.New(rand.NewSource(rng.Int63()))
		eg.Go(func() error {
			j := f.makeJob(cp, subrng, resCh)
			return j.Fuzz(ectx)
		})
	}

	eg.Go(func() error {
		return handleResults(ectx, fcs, resCh)
	})
	err := eg.Wait()
	return fcs, err
}

func (f *Fuzzer) makeJob(cp model.Corpus, subrng *rand.Rand, resCh chan model.Subject) *job {
	return &job{
		Corpus:        cp,
		Driver:        f.Driver,
		SubjectCycles: f.SubjectCycles,
		Pathset:       f.Paths,
		Rng:           subrng,
		ResCh:         resCh,
	}
}

func (f *Fuzzer) corpusChunks() []model.Corpus {
	nchunks := len(f.Plan.Corpus)
	if 0 < f.FuzzWorkers && f.FuzzWorkers < nchunks {
		nchunks = f.FuzzWorkers
	}
	return f.Plan.Corpus.Chunks(nchunks)
}
