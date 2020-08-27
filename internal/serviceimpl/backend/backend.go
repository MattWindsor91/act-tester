// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

// Package backend contains style-to-backend resolution.
package backend

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/MattWindsor91/act-tester/internal/serviceimpl/backend/herdtools/litmus"

	"github.com/MattWindsor91/act-tester/internal/serviceimpl/backend/herdtools/herd"

	"github.com/MattWindsor91/act-tester/internal/model/recipe"

	"github.com/MattWindsor91/act-tester/internal/serviceimpl/backend/herdtools"

	"github.com/MattWindsor91/act-tester/internal/model/job"

	"github.com/MattWindsor91/act-tester/internal/subject/obs"

	"github.com/MattWindsor91/act-tester/internal/model/service"
	"github.com/MattWindsor91/act-tester/internal/stage/lifter"
	"github.com/MattWindsor91/act-tester/internal/stage/mach/runner"
)

var (
	// ErrNil occurs when the compiler we try to resolve is nil.
	ErrNil = errors.New("compiler nil")
	// ErrUnknownStyle occurs when we ask the resolver for a backend style of which it isn't aware.
	ErrUnknownStyle = errors.New("unknown backend style")

	// BResolve is a pre-populated compiler resolver.
	BResolve = Resolver{Backends: map[string]Backend{
		"herd": herdtools.Backend{
			DefaultRun: service.RunInfo{Cmd: "herd7"},
			Impl:       herd.Herd{},
		},
		"litmus": herdtools.Backend{
			DefaultRun: service.RunInfo{Cmd: "litmus7"},
			Impl:       litmus.Litmus{},
		},
	}}
)

// Backend contains the various interfaces that a backend can implement.
type Backend interface {
	lifter.SingleLifter
	runner.ObsParser
}

// Inspector maps compiler styles to compilers.
type Resolver struct {
	// Compilers is the raw map from style strings to backend runners.
	Backends map[string]Backend
}

// Lift delegates lifting to the appropriate maker for j.
func (r *Resolver) Lift(ctx context.Context, j job.Lifter, errw io.Writer) (recipe.Recipe, error) {
	bi, err := r.Get(j.Backend)
	if err != nil {
		return recipe.Recipe{}, err
	}
	return bi.Lift(ctx, j, errw)
}

// ParseObs delegates observation parsing to the appropriate implementation for the backend referenced by b.
func (r *Resolver) ParseObs(ctx context.Context, b *service.Backend, rd io.Reader, o *obs.Obs) error {
	bi, err := r.Get(b)
	if err != nil {
		return err
	}
	return bi.ParseObs(ctx, b, rd, o)
}

// Get tries to look up the backend specified by b in this resolver.
func (r *Resolver) Get(b *service.Backend) (Backend, error) {
	if r == nil {
		return nil, ErrNil
	}

	sstr := b.Style.String()
	bi, ok := r.Backends[sstr]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownStyle, sstr)
	}
	return bi, nil
}
