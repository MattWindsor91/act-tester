// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package compilation

import (
	"time"

	"github.com/MattWindsor91/act-tester/internal/model/status"
)

// Result is the base structure for things that represent the result of an external process.
type Result struct {
	// Time is the time at which the process commenced.
	Time time.Time `json:"time,omitempty"`

	// Duration is the rough duration of the process.
	Duration time.Duration `json:"duration,omitempty"`

	// Status is the status of the process.
	Status status.Status `json:"status"`
}