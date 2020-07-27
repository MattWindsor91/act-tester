// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package mach_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MattWindsor91/act-tester/internal/stage/mach/compiler"
	"github.com/MattWindsor91/act-tester/internal/stage/mach/runner"
	"github.com/MattWindsor91/act-tester/internal/stage/mach/timeout"

	"github.com/MattWindsor91/act-tester/internal/stage/mach"
)

// ExampleQuantitySet_Log is a testable example for Log.
func ExampleQuantitySet_Log() {
	qs := mach.QuantitySet{
		Compiler: compiler.QuantitySet{
			Timeout: timeout.Timeout(1 * time.Minute),
		},
		Runner: runner.QuantitySet{
			Timeout:  timeout.Timeout(2 * time.Minute),
			NWorkers: 1,
		},
	}

	l := log.New(os.Stdout, "", 0)
	qs.Log(l)

	// Output:
	// [Compiler]
	// timeout at 1m0s
	// [Runner]
	// running across 1 worker
	// timeout at 2m0s
}

// TestQuantitySet_Override tests Override against some cases.
func TestQuantitySet_Override(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		old, new, want mach.QuantitySet
	}{
		"all-old": {
			new: mach.QuantitySet{
				Compiler: compiler.QuantitySet{
					Timeout: timeout.Timeout(4 * time.Second),
				},
				Runner: runner.QuantitySet{
					Timeout:  timeout.Timeout(98 * time.Minute),
					NWorkers: 27,
				},
			},
			old: mach.QuantitySet{},
			want: mach.QuantitySet{
				Compiler: compiler.QuantitySet{
					Timeout: timeout.Timeout(4 * time.Second),
				},
				Runner: runner.QuantitySet{
					Timeout:  timeout.Timeout(98 * time.Minute),
					NWorkers: 27,
				},
			},
		},
		"all-new": {
			old: mach.QuantitySet{},
			new: mach.QuantitySet{
				Compiler: compiler.QuantitySet{
					Timeout: timeout.Timeout(42 * time.Second),
				},
				Runner: runner.QuantitySet{
					Timeout:  timeout.Timeout(1 * time.Minute),
					NWorkers: 42,
				},
			},
			want: mach.QuantitySet{
				Compiler: compiler.QuantitySet{
					Timeout: timeout.Timeout(42 * time.Second),
				},
				Runner: runner.QuantitySet{
					Timeout:  timeout.Timeout(1 * time.Minute),
					NWorkers: 42,
				},
			},
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := c.old
			got.Override(c.new)

			assert.Equal(t, c.want, got, "quantity set override mismatch")
		})
	}
}
