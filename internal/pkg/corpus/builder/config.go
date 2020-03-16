// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package builder

import "github.com/MattWindsor91/act-tester/internal/pkg/corpus"

// Config is a configuration for a Builder.
type Config struct {
	// Init is the initial corpus.
	// If nil, the Builder starts with a new corpus with capacity equal to NReqs.
	// Otherwise, it copies this corpus.
	Init corpus.Corpus

	// Manifest gives us the name of the task and the number of requests in it.
	Manifest

	// Obs is the observer to notify as the builder performs various tasks.
	Obs Observer
}