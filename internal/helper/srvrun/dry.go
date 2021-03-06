// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package srvrun

import (
	"context"
	"io"
	"time"

	"github.com/c4-project/c4t/internal/model/service"
)

// DryRunner is a runner that just outputs all commands to be run to itself.
type DryRunner struct {
	io.Writer
}

// WithStdout just returns the same runner, ignoring the override.
func (d DryRunner) WithStdout(_ io.Writer) service.Runner {
	return d
}

// WithStderr just returns the same runner, ignoring the override.
func (d DryRunner) WithStderr(_ io.Writer) service.Runner {
	return d
}

// WithGrace just returns the same runner, ignoring the override.
func (d DryRunner) WithGrace(_ time.Duration) service.Runner {
	return d
}

// Run dumps r to the dry runner's writer.
func (d DryRunner) Run(_ context.Context, r service.RunInfo) error {
	_, err := io.WriteString(d, r.String()+"\n")
	return err
}
