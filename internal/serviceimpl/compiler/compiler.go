// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

// Package compiler contains style-to-compiler resolution.
package compiler

import (
	"context"
	"errors"
	"fmt"

	"github.com/c4-project/c4t/internal/id"

	"github.com/c4-project/c4t/internal/stage/mach/interpreter"

	"github.com/c4-project/c4t/internal/helper/stringhelp"

	"github.com/c4-project/c4t/internal/model/service/compiler/optlevel"

	"github.com/c4-project/c4t/internal/serviceimpl/compiler/gcc"

	mdl "github.com/c4-project/c4t/internal/model/service/compiler"

	"github.com/c4-project/c4t/internal/model/service"
)

var (
	// ErrNil occurs when the compiler we try to resolve is nil.
	ErrNil = errors.New("compiler nil")
	// ErrUnknownStyle occurs when we ask the resolver for a compiler style of which it isn't aware.
	ErrUnknownStyle = errors.New("unknown compiler style")

	// CResolve is a pre-populated compiler resolver.
	CResolve = Resolver{Compilers: map[id.ID]Compiler{
		id.CStyleGCC: gcc.GCC{
			DefaultRunInfo: service.RunInfo{Cmd: "gcc", Args: []string{"-pthread", "-std=gnu11"}},
			AltCommands: []string{
				// non-exhaustive, add more as we need them
				"clang",
			},
		},
	}}
)

// Compiler contains the various interfaces that a compiler can implement.
type Compiler interface {
	// Probe uses sr to probe for copies of a particular compiler class with id classId, adding them to target.
	Probe(ctx context.Context, sr service.Runner, classId id.ID, target mdl.ConfigMap) error

	mdl.Inspector
	interpreter.Driver
}

//go:generate mockery --name=Compiler

// Resolver maps compiler styles to compilers.
type Resolver struct {
	// Compilers is the raw map from style strings to compiler runners.
	Compilers map[id.ID]Compiler
}

// Get tries to look up the compiler specified by nc in this resolver.
func (r *Resolver) Get(c *mdl.Compiler) (Compiler, error) {
	if c == nil {
		return nil, ErrNil
	}
	cp, ok := r.Compilers[c.Style]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownStyle, c.Style)
	}
	return cp, nil
}

// DefaultOptLevels gets the default optimisation levels for the compiler described by c.
func (r *Resolver) DefaultOptLevels(c *mdl.Compiler) (stringhelp.Set, error) {
	cp, err := r.Get(c)
	if err != nil {
		return nil, err
	}
	return cp.DefaultOptLevels(c)
}

// OptLevels gets information about all available optimisation levels for the compiler described by c.
func (r *Resolver) OptLevels(c *mdl.Compiler) (map[string]optlevel.Level, error) {
	cp, err := r.Get(c)
	if err != nil {
		return nil, err
	}
	return cp.OptLevels(c)
}

// DefaultMOpts gets the default machine-specific optimisation profiles for the compiler described by c.
func (r *Resolver) DefaultMOpts(c *mdl.Compiler) (stringhelp.Set, error) {
	cp, err := r.Get(c)
	if err != nil {
		return nil, err
	}
	return cp.DefaultMOpts(c)
}

// RunCompiler runs the compiler specified by nc on job j, using this resolver to map the style to a concrete compiler.
func (r *Resolver) RunCompiler(ctx context.Context, j mdl.Job, sr service.Runner) error {
	cp, err := r.Get(&j.Compiler.Compiler)
	if err != nil {
		return err
	}
	return cp.RunCompiler(ctx, j, sr)
}

func (r *Resolver) Probe(ctx context.Context, sr service.Runner) (mdl.ConfigMap, error) {
	// As an educated guess, assume every class has one spec.
	target := make(mdl.ConfigMap, len(r.Compilers))
	for cid, class := range r.Compilers {
		if err := class.Probe(ctx, sr, cid, target); err != nil {
			return nil, err
		}
	}
	return target, nil
}
