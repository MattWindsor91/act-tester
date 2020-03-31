// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package view

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// LogTopError logs err if non-nil, in a 'fatal top-level error' sort of way.
func LogTopError(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		// Assume that a top-level cancellation is user-specified.
		return
	}

	var perr *exec.ExitError
	if errors.As(err, &perr) {
		_, _ = fmt.Fprintln(os.Stderr, "A child process encountered an error:")
		_, _ = fmt.Fprintln(os.Stderr, err)
		_, _ = fmt.Fprintln(os.Stderr, "Any captured stderr follows.")
		_, _ = fmt.Fprintln(os.Stderr, perr.Stderr)
		os.Exit(perr.ExitCode())
		return
	}

	_, _ = fmt.Fprintln(os.Stderr, "A fatal error has occurred:")
	_, _ = fmt.Fprintln(os.Stderr, err)

	os.Exit(1)
}