// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package fuzzer

import (
	"context"

	"github.com/MattWindsor91/act-tester/internal/model/job"
	"github.com/MattWindsor91/act-tester/internal/model/litmus"
)

// Driver groups the interfaces used to 'drive' parts of the fuzzer.
type Driver interface {
	SingleFuzzer
	litmus.StatDumper
}

// AggregateDriver is a driver that delegates the interface responsibilities to separate implementations.
type AggregateDriver struct {
	// Single is a single-job fuzzer.
	Single SingleFuzzer
	// Stat is a stat dumper.
	Stat litmus.StatDumper
}

// Fuzz delegates to the fuzzer.
func (a AggregateDriver) Fuzz(ctx context.Context, job job.Fuzzer) error {
	return a.Single.Fuzz(ctx, job)
}

// DumpStats delegates to the stat dumper.
func (a AggregateDriver) DumpStats(ctx context.Context, s *litmus.Statset, path string) error {
	return a.Stat.DumpStats(ctx, s, path)
}
