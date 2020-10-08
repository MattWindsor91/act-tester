// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package coverage

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/1set/gut/yos"

	"github.com/MattWindsor91/act-tester/internal/plan"
)

// Maker contains state used by the coverage testbed maker.
type Maker struct {
	// outDir is the name of the output directory.
	outDir string

	// profiles contains the map of profiles available to the coverage testbed maker.
	profiles map[string]Profile

	// qs is the calculated quantity set for the coverage testbed maker.
	qs QuantitySet
}

// NewMaker constructs a new coverage testbed maker.
func NewMaker(profiles map[string]Profile, opts ...Option) (*Maker, error) {
	m := &Maker{profiles: profiles}
	if err := Options(opts...)(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Maker) Run(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	buckets := m.qs.Buckets()
	if buckets == nil {
		return nil, errors.New("bucket calculation failed")
	}

	if err := m.prepare(buckets); m != nil {
		return nil, err
	}

	// for now
	return p, nil
}

func (m *Maker) prepare(buckets map[string]int) error {
	for pname := range m.profiles {
		for suffix := range buckets {
			if err := yos.MakeDir(m.bucketDir(pname, suffix)); err != nil {
				return fmt.Errorf("preparing directory for profile %q bucket %q: %w", pname, suffix, err)
			}
		}
	}
	return nil
}

func (m *Maker) bucketDir(pname string, suffix string) string {
	return filepath.Join(m.outDir, pname, suffix)
}