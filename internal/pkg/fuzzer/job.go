package fuzzer

import (
	"context"
	"math/rand"

	"github.com/MattWindsor91/act-tester/internal/pkg/model"
)

// job contains state for a single fuzzer batch-job.
type job struct {
	// Corpus contains the corpus for which this job is responsible.
	Corpus model.Corpus

	// Driver is the low-level fuzzer.
	Driver SingleFuzzer

	// SubjectCycles is the number of times each subject should be fuzzed.
	SubjectCycles int

	// Pathset points to the pathset to use to work out where to store fuzz output.
	Pathset *Pathset

	// Rng is the random number generator to use for fuzz seeds.
	Rng *rand.Rand

	// ResCh is the channel onto which each fuzzed subject should be sent.
	ResCh chan<- model.Subject
}

func (j *job) Fuzz(ctx context.Context) error {
	for i := range j.Corpus {
		if err := j.fuzzSubject(ctx, &j.Corpus[i]); err != nil {
			return nil
		}
	}
	return nil
}

// fuzzSubject fuzzes subject s with a seed generated by rng, storing according to ps and announcing progress on bar.
func (j *job) fuzzSubject(ctx context.Context, s *model.Subject) error {
	for i := 0; i < j.SubjectCycles; i++ {
		if err := j.fuzzCycle(ctx, s, i); err != nil {
			return err
		}
	}
	return nil
}

func (j *job) fuzzCycle(ctx context.Context, s *model.Subject, cycle int) error {
	outp, tracep := j.Pathset.OnSubject(s.Name, cycle)
	if err := j.Driver.FuzzSingle(j.Rng.Int31(), s.Litmus, outp, tracep); err != nil {
		return err
	}
	s2 := model.Subject{
		Name:       CycledName(s.Name, cycle),
		OrigLitmus: s.Litmus,
		Litmus:     outp,
		TracePath:  tracep,
	}
	if err := j.sendSubject(ctx, s2); err != nil {
		return err
	}
	return nil
}

// sendSubject tries to send s down this job's result channel.
func (j *job) sendSubject(ctx context.Context, s model.Subject) error {
	select {
	case j.ResCh <- s:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
