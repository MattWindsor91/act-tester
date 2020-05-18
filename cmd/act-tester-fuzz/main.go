// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package main

import (
	"os"

	"github.com/MattWindsor91/act-tester/internal/app/fuzz"

	"github.com/MattWindsor91/act-tester/internal/view"
)

func main() {
	app := fuzz.App(os.Stdout, os.Stderr)
	view.LogTopError(app.Run(os.Args))
}
