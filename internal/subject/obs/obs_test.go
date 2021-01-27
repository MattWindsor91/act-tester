// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package obs_test

import (
	"fmt"
	"testing"

	"github.com/c4-project/c4t/internal/subject/status"

	"github.com/c4-project/c4t/internal/subject/obs"

	"github.com/c4-project/c4t/internal/helper/testhelp"
)

// Used to avoid compiler-optimising-out of the benchmark below.
var stat status.Status

// BenchmarkObs_Status benchmarks Obs.Status.
func BenchmarkObs_Status(b *testing.B) {
	cases := map[string]*obs.Obs{
		"empty":   {},
		"undef":   {Flags: obs.Undef},
		"sat":     {Flags: obs.Sat},
		"unsat":   {Flags: obs.Unsat},
		"e-sat":   {Flags: obs.Sat | obs.Exist},
		"e-unsat": {Flags: obs.Unsat | obs.Exist},
	}

	for name, c := range cases {
		c := c
		b.Run(name, func(b *testing.B) {
			s := stat
			for i := 0; i < b.N; i++ {
				s = c.Status()
			}
			stat = s
		})

	}

}

// ExampleObs_Status is a testable example for Obs.Status.
func ExampleObs_Status() {
	fmt.Println("empty:  ", (&obs.Obs{}).Status())
	fmt.Println("undef:  ", (&obs.Obs{Flags: obs.Undef}).Status())
	fmt.Println("sat:    ", (&obs.Obs{Flags: obs.Sat}).Status())
	fmt.Println("unsat:  ", (&obs.Obs{Flags: obs.Unsat}).Status())
	fmt.Println("e-sat:  ", (&obs.Obs{Flags: obs.Sat | obs.Exist}).Status())
	fmt.Println("e-unsat:", (&obs.Obs{Flags: obs.Unsat | obs.Exist}).Status())

	// output:
	// empty:   Flagged
	// undef:   Flagged
	// sat:     Ok
	// unsat:   Flagged
	// e-sat:   Flagged
	// e-unsat: Ok
}

func TestObs_TOML_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := map[string]obs.Obs{
		"empty":         {},
		"undef-nostate": {Flags: obs.Undef},
		"multiple-flags": {
			Flags: obs.Sat | obs.Undef,
			States: []obs.State{
				{"x": "27", "y": "53"},
				{"x": "27", "y": "42"},
			},
			Witnesses: []obs.State{
				{"x": "27", "y": "53"},
			},
		},
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testhelp.TestTomlRoundTrip(t, want, "Obs")
		})
	}
}
