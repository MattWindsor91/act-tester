// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package forward

import (
	"encoding/json"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"
)

// Observer wraps a JSON encoder, lifting it to an Observer that sends JSON-encoded Forwards.
type Observer struct {
	*json.Encoder
}

// OnBuildStart sends a build message through this Observer's encoder.
func (o *Observer) OnBuild(m builder.Message) {
	o.forwardHandlingError(Forward{Build: &m})
}

// Error forwards err to this Observer's encoder.
func (o *Observer) Error(err error) {
	_ = o.forward(Forward{Error: err.Error()})
}

func (o *Observer) forwardHandlingError(f Forward) {
	if err := o.forward(f); err != nil {
		o.Error(err)
	}
}

func (o *Observer) forward(f Forward) error {
	return o.Encode(f)
}