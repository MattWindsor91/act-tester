// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package csvdump_test

import (
	"context"
	"os"

	"github.com/c4-project/c4t/internal/stage/analyser/csvdump"

	"github.com/c4-project/c4t/internal/plan"
	"github.com/c4-project/c4t/internal/plan/analysis"
)

// NB: the below CSV is likely to change as the plan mock changes.
// At time of writing, the mock referred to compilers not in the plan, for instance.

// TODO(@MattWindsor91): add stages to the mock plan!

// ExampleStageWriter_OnAnalysis is a testable example for OnAnalysis.
func ExampleStageWriter_OnAnalysis() {
	an, _ := analysis.Analyse(context.Background(), plan.Mock())

	sw := csvdump.NewStageWriter(os.Stdout)
	sw.OnAnalysis(*an)

	// Output:
	// Stage,Start,End,Duration
}
