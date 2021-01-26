// Copyright (c) 2021 Matt Windsor and contributors
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package mutation_test

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/c4-project/c4t/internal/subject/status"

	"github.com/c4-project/c4t/internal/model/id"
	"github.com/c4-project/c4t/internal/mutation"
	"github.com/c4-project/c4t/internal/subject/compilation"
)

// ExampleStatset_Reset is a runnable example for Statset.Reset.
func ExampleStatset_Reset() {
	var s mutation.Statset
	s.Reset()

	fmt.Println("by-mutant nil:", s.ByMutant == nil, "len:", len(s.ByMutant))

	// Output:
	// by-mutant nil: false len: 0
}

// ExampleStatset_AddAnalysis is a runnable example for AddAnalysis.
func ExampleStatset_AddAnalysis() {
	var s mutation.Statset
	s.AddAnalysis(mutation.Analysis{
		27: mutation.MutantAnalysis{
			{
				NumHits: 0,
				Status:  status.Ok,
				HitBy:   compilation.Name{SubjectName: "smooth", CompilerID: id.FromString("criminal")},
			},
			{
				NumHits: 2,
				Status:  status.RunFail,
				HitBy:   compilation.Name{SubjectName: "marco", CompilerID: id.FromString("polo")},
			},
			{
				NumHits: 4,
				Status:  status.Flagged,
				HitBy:   compilation.Name{SubjectName: "mint", CompilerID: id.FromString("polo")},
			},
		},
		53: mutation.MutantAnalysis{
			{
				NumHits: 0,
				Status:  status.Filtered,
				HitBy:   compilation.Name{SubjectName: "marco", CompilerID: id.FromString("polo")},
			},
		},
	})

	fmt.Println("27 selected:", s.ByMutant[27].Selections, "hit:", s.ByMutant[27].Hits, "killed:", s.ByMutant[27].Kills)
	fmt.Println("53 selected:", s.ByMutant[53].Selections, "hit:", s.ByMutant[53].Hits, "killed:", s.ByMutant[53].Kills)

	// Output:
	// 27 selected: 3 hit: 6 killed: 1
	// 53 selected: 1 hit: 0 killed: 0
}

// ExampleStatset_DumpCSV is a runnable example for Statset.DumpCSV.
func ExampleStatset_DumpCSV() {
	_ = (&mutation.Statset{
		ByMutant: map[mutation.Mutant]mutation.MutantStatset{
			2:  {Selections: 1, Hits: 0, Kills: 0, Statuses: map[status.Status]uint64{status.Filtered: 1}},
			42: {Selections: 10, Hits: 1, Kills: 0, Statuses: map[status.Status]uint64{status.Ok: 9, status.CompileTimeout: 1}},
			53: {Selections: 20, Hits: 400, Kills: 15, Statuses: map[status.Status]uint64{status.Flagged: 15, status.CompileFail: 3, status.RunFail: 2}},
		},
	}).DumpCSV(csv.NewWriter(os.Stdout), "localhost")

	// Output:
	// localhost,2,1,0,0,0,1,0,0,0,0,0
	// localhost,42,10,1,0,9,0,0,0,1,0,0
	// localhost,53,20,400,15,0,0,15,3,0,2,0
}
