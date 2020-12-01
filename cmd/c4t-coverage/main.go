// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package main

import (
	"os"

	"github.com/MattWindsor91/c4t/internal/app/coverage"

	"github.com/MattWindsor91/c4t/internal/ux"
)

func main() {
	ux.LogTopError(coverage.App(os.Stdout, os.Stderr).Run(os.Args))
}