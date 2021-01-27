// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package main

import (
	"os"

	"github.com/c4-project/c4t/internal/app/setc"
	"github.com/c4-project/c4t/internal/ux"
)

func main() {
	ux.LogTopError(setc.App(os.Stdout, os.Stderr).Run(os.Args))
}
