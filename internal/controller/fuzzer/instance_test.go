// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package fuzzer_test

import (
	"context"
	"math/rand"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/MattWindsor91/act-tester/internal/model/litmus/mocks"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"

	"github.com/MattWindsor91/act-tester/internal/controller/fuzzer"
	"github.com/MattWindsor91/act-tester/internal/model/subject"
	"golang.org/x/sync/errgroup"
)

// TestJob_Fuzz tests various aspects of a job fuzz.
func TestJob_Fuzz(t *testing.T) {
	resCh := make(chan builder.Request)

	var md mocks.StatDumper

	j := fuzzer.Instance{
		Subject:       subject.Named{Name: "foo"},
		Driver:        fuzzer.NopFuzzer{},
		StatDumper:    &md,
		SubjectCycles: 10,
		Pathset:       fuzzer.NewPathset("test"),
		Rng:           rand.New(rand.NewSource(0)),
		ResCh:         resCh,
	}

	for i := 0; i < 10; i++ {
		i := i
		wname := path.Join("test", "litmus", fuzzer.SubjectCycle{Name: "foo", Cycle: i}.String()+".litmus")
		md.On("DumpStats", mock.Anything, mock.Anything, wname).Return(nil).Once()
	}

	eg, ectx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		return j.Fuzz(ectx)
	})
	eg.Go(func() error {
		for i := 0; i < 10; i++ {
			select {
			case r := <-resCh:
				// TODO(@MattWindsor91): other checks
				wname := fuzzer.SubjectCycle{Name: "foo", Cycle: i}.String()
				if r.Name != wname {
					t.Errorf("wrong fuzz result name: got=%q, want=%q", r.Name, wname)
				}
			case <-ectx.Done():
				return ectx.Err()
			}
		}
		return nil
	})
	assert.NoError(t, eg.Wait(), "unexpected errgroup error")

	md.AssertExpectations(t)
}