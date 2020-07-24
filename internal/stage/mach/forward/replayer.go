// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package forward

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"
)

var ErrRemote = errors.New("remote error")

// Replayer coordinates reading forwarded builder-status messages from a JSON decoder and replaying them to an observer.
type Replayer struct {
	// Decoder is the decoder on which we are listening for messages to replay.
	Decoder *json.Decoder

	// Observers is the set of observers to which we are forwarding observations.
	Observers []builder.Observer
}

// Run runs the replayer.
func (r *Replayer) Run(ctx context.Context) error {
	for {
		if err := checkClose(ctx); err != nil {
			return err
		}

		var f Forward
		if err := r.Decoder.Decode(&f); err != nil {
			// EOF is entirely expected at some point.
			if errors.Is(err, io.EOF) {
				return ctx.Err()
			}
			return fmt.Errorf("while decoding updates: %w", err)
		}

		if err := r.forwardToObs(f); err != nil {
			return fmt.Errorf("while forwarding updates: %w", err)
		}
	}
}

func (r *Replayer) forwardToObs(f Forward) error {
	switch {
	case f.Error != "":
		return fmt.Errorf("%w: %s", ErrRemote, f.Error)
	case f.Build != nil:
		builder.OnBuild(*f.Build, r.Observers...)
		return nil
	default:
		return errors.New("received forward with nothing present")
	}
}

func checkClose(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}