// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package compiler

import (
	"log"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"
)

// Option is the type of options to the compiler sub-stage constructor.
type Option func(*Compiler) error

// Options applies each option in opts in turn.
func Options(opts ...Option) Option {
	return func(c *Compiler) error {
		for _, o := range opts {
			if err := o(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// LogTo sets the runner's logger to l.
func LogTo(l *log.Logger) Option {
	// TODO(@MattWindsor91): as elsewhere, logging should be replaced with observing
	return func(c *Compiler) error {
		// Logger ensuring is done after all options are processed
		c.l = l
		return nil
	}
}

// ObserveWith adds each observer in obs to the runner's observer list.
func ObserveWith(obs ...builder.Observer) Option {
	return func(c *Compiler) error {
		var err error
		c.observers, err = builder.AppendObservers(c.observers, obs...)
		return err
	}
}

// OverrideQuantities overrides this runner's quantities with qs.
func OverrideQuantities(qs QuantitySet) Option {
	return func(c *Compiler) error {
		c.quantities.Override(qs)
		return nil
	}
}