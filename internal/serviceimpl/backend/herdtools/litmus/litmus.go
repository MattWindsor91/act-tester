// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package litmus contains the parts of a Herdtools backend specific to litmus7.
package litmus

import (
	"context"
	"fmt"
	"io"

	"github.com/MattWindsor91/act-tester/internal/model/service/backend"

	"github.com/MattWindsor91/act-tester/internal/model/service"
)

// Litmus describes the parts of a backend invocation that are specific to Litmus.
type Litmus struct{}

// LiftExe runs litmus in a mode that generates files compilable into an executable.
func (l Litmus) LiftExe(ctx context.Context, j backend.LiftJob, r service.RunInfo, x service.Runner) error {
	i := Instance{Job: j, RunInfo: r, Runner: x}
	return i.Run(ctx)
}

func litmusCommonArgs(j backend.LiftJob) ([]string, error) {
	carch, err := lookupArch(j.Arch)
	if err != nil {
		return nil, fmt.Errorf("when looking up -carch: %w", err)
	}

	return []string{"-carch", carch, "-c11", "true", j.In.Litmus.Path}, nil
}

// LiftStandalone runs litmus in standalone mode.
// It currently doesn't do the same patching as LiftExe does.
func (l Litmus) LiftStandalone(ctx context.Context, j backend.LiftJob, r service.RunInfo, x service.Runner, w io.Writer) error {
	// TODO(@MattWindsor91): ideally we'd do patching here too, but that'll be complicated to do in a standalone manner.
	args, err := litmusCommonArgs(j)
	if err != nil {
		return err
	}
	r.Override(service.RunInfo{Args: args})
	return x.WithStdout(w).Run(ctx, r)
}
