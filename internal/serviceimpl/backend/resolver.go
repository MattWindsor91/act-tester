// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

// Package resolver contains the backend resolver.
package backend

import (
	"errors"
	"fmt"

	"github.com/c4-project/c4t/internal/model/id"

	backend2 "github.com/c4-project/c4t/internal/model/service/backend"

	"github.com/c4-project/c4t/internal/model/service"
	"github.com/c4-project/c4t/internal/serviceimpl/backend/delitmus"
	"github.com/c4-project/c4t/internal/serviceimpl/backend/herdstyle"
	"github.com/c4-project/c4t/internal/serviceimpl/backend/herdstyle/herd"
	"github.com/c4-project/c4t/internal/serviceimpl/backend/herdstyle/litmus"
	"github.com/c4-project/c4t/internal/serviceimpl/backend/herdstyle/rmem"
)

var (
	// ErrNil occurs when the backend we try to resolve is nil.
	ErrNil = errors.New("backend nil")
	// ErrUnknownStyle occurs when we ask the resolver for a backend style of which it isn't aware.
	ErrUnknownStyle = errors.New("unknown backend style")

	herdArches   = []id.ID{id.ArchC, id.ArchAArch64, id.ArchArm, id.ArchX8664, id.ArchX86, id.ArchPPC}
	litmusArches = []id.ID{id.ArchC, id.ArchAArch64, id.ArchArm, id.ArchX8664, id.ArchX86, id.ArchPPC}
	// TODO(@MattWindsor91): rmem supports more than this, but needs more work on sanitising/model selection
	rmemArches = []id.ID{id.ArchAArch64}

	// Resolve is a pre-populated backend resolver.
	Resolve = Resolver{Backends: map[string]func(r *service.RunInfo) backend2.Backend{
		"delitmus": func(*service.RunInfo) backend2.Backend { return delitmus.Delitmus{} },
		"herdtools.herd": herdstyle.Backend{
			OptCapabilities: 0,
			Arches:          herdArches,
			RunInfo:         service.RunInfo{Cmd: "herd7"},
			Impl:            herd.Herd{},
		}.Instantiate,
		"herdtools.litmus": herdstyle.Backend{
			OptCapabilities: backend2.CanProduceExe,
			Arches:          litmusArches,
			RunInfo:         service.RunInfo{Cmd: "litmus7"},
			Impl:            litmus.Litmus{},
		}.Instantiate,
		"rmem": herdstyle.Backend{
			OptCapabilities: backend2.CanLiftLitmus,
			Arches:          rmemArches,
			RunInfo:         service.RunInfo{Cmd: "rmem"},
			Impl:            rmem.Rmem{},
		}.Instantiate,
	}}
)

// Resolver maps backend styles to backends.
type Resolver struct {
	// Backends is the raw map from style strings to backend constructors.
	Backends map[string]func(ri *service.RunInfo) backend2.Backend
}

// Resolve tries to look up the backend specified by b in this resolver.
func (r *Resolver) Resolve(b *backend2.Spec) (backend2.Backend, error) {
	if r == nil {
		return nil, ErrNil
	}

	sstr := b.Style.String()
	bi, ok := r.Backends[sstr]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownStyle, sstr)
	}
	return bi(b.Run), nil
}
