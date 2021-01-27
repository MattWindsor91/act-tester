// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package iohelp

import (
	"io"
	"io/ioutil"
)

// EnsureWriter passes through w if non-nil, or supplies ioutil.Discard otherwise.
func EnsureWriter(w io.Writer) io.Writer {
	if w == nil {
		return ioutil.Discard
	}
	return w
}
