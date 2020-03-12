// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package fuzzer

import (
	"context"
	"log"
	"reflect"

	"github.com/MattWindsor91/act-tester/internal/pkg/corpus"

	"github.com/MattWindsor91/act-tester/internal/pkg/subject"

	"github.com/MattWindsor91/act-tester/internal/pkg/plan"
)

// SubjectPather is the interface of things that serve file-paths for subject outputs during a fuzz batch.
type SubjectPather interface {
	// Prepare sets up the directories ready to serve through SubjectPaths.
	Prepare() error

	// SubjectPaths gets the litmus and trace file paths for the subject/cycle pair sc.
	SubjectPaths(sc SubjectCycle) subject.FuzzFileset
}

// Config represents the configuration that goes into a batch fuzzer run.
type Config struct {
	// Driver holds the single-file fuzzer that the fuzzer is going to use.
	Driver SingleFuzzer

	// Logger is the logger to use while fuzzing.  It may be nil, which silences logging.
	Logger *log.Logger

	// Observer observes the fuzzer's progress across a corpus.
	Observer corpus.BuilderObserver

	// Paths contains the path set for things generated by this fuzzer.
	Paths SubjectPather

	// Quantities sets the quantities for this batch fuzzer run.
	Quantities QuantitySet
}

// QuantitySet represents the part of a configuration that holds various tunable parameters for the batch runner.
type QuantitySet struct {
	// CorpusSize is the sampling size for the corpus after fuzzing.
	// It has a similar effect to CorpusSize in planner.Planner.
	CorpusSize int

	// SubjectCycles is the number of times to fuzz each file.
	SubjectCycles int
}

// Override substitutes any quantities in new that are non-zero for those in this set.
func (q *QuantitySet) Override(new QuantitySet) {
	qv := reflect.ValueOf(q).Elem()
	nv := reflect.ValueOf(new)

	nf := nv.NumField()
	for i := 0; i < nf; i++ {
		k := nv.Field(i).Int()
		if k != 0 {
			qv.Field(i).SetInt(k)
		}
	}
}

// Run runs a fuzzer configured by this config.
func (c *Config) Run(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	f, err := New(c, p)
	if err != nil {
		return nil, err
	}
	return f.Fuzz(ctx)
}
