// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package planner_test

import (
	"context"
	"math/rand"
	"sort"
	"testing"

	"github.com/1set/gut/ystring"

	"github.com/stretchr/testify/assert"

	"github.com/MattWindsor91/act-tester/internal/model/service"

	"github.com/stretchr/testify/require"

	"github.com/MattWindsor91/act-tester/internal/controller/planner"

	"github.com/MattWindsor91/act-tester/internal/helper/stringhelp"
	"github.com/MattWindsor91/act-tester/internal/model/compiler"
	"github.com/MattWindsor91/act-tester/internal/model/compiler/optlevel"
	"github.com/MattWindsor91/act-tester/internal/model/id"
	"github.com/stretchr/testify/mock"
)

// TestCompilerPlanner_Plan tests the happy path of a compiler planner using copious amounts of mocking.
func TestCompilerPlanner_Plan(t *testing.T) {
	var (
		mi mockInspector
		ml mockCompilerLister
		mo mockCompilerObserver
	)

	rng := rand.New(rand.NewSource(0))
	ctx := context.Background()
	mid := id.FromString("localhost")

	cfgs := map[string]compiler.Config{
		"gcc": {
			Style: id.CStyleGCC,
			Arch:  id.ArchArmCortexA72,
		},
		"clang": {
			Style: id.CStyleGCC,
			Arch:  id.ArchArm8,
			Run: &service.RunInfo{
				Cmd: "clang",
			},
			MOpt: &optlevel.Selection{
				Enabled:  []string{"march=armv8-a"},
				Disabled: []string{"march=armv7-a"},
			},
			Opt: &optlevel.Selection{
				Enabled:  []string{"1"},
				Disabled: []string{"fast"},
			},
		},
	}
	ncfgs := len(cfgs)

	dls := stringhelp.NewSet("0", "2", "fast")
	dms := stringhelp.NewSet("march=armv7-a")

	ols := map[string]optlevel.Level{
		"0": {
			Optimises:       false,
			Bias:            optlevel.BiasDebug,
			BreaksStandards: false,
		},
		"1": {
			Optimises:       true,
			Bias:            optlevel.BiasSize,
			BreaksStandards: false,
		},
		"2": {
			Optimises:       true,
			Bias:            optlevel.BiasSpeed,
			BreaksStandards: false,
		},
		"fast": {
			Optimises:       true,
			Bias:            optlevel.BiasSpeed,
			BreaksStandards: true,
		},
	}

	ml.On("ListCompilers", ctx, mid).Return(cfgs, nil).Once()

	mo.On("OnCompilerPlanStart", len(cfgs)).Return().Once()
	mo.On("OnCompilerPlanFinish").Return().Once()

	mi.On("DefaultMOpts", mock.Anything).Return(dms, nil).Times(ncfgs)
	mi.On("DefaultOptLevels", mock.Anything).Return(dls, nil).Times(ncfgs)
	mi.On("OptLevels", mock.Anything).Return(ols, nil).Times(ncfgs)

	keys, _ := stringhelp.MapKeys(cfgs)
	sort.Strings(keys)

	mo.On("OnCompilerPlan", mock.MatchedBy(func(c compiler.Named) bool {
		cs := c.ID.String()
		i := sort.SearchStrings(keys, cs)
		return i < ncfgs && keys[i] == cs
	})).Return().Times(ncfgs)

	cp := planner.CompilerPlanner{
		Lister:    &ml,
		Inspector: &mi,
		Observers: []planner.CompilerObserver{&mo},
		MachineID: mid,
		Rng:       rng,
	}

	cs, err := cp.Plan(ctx)
	require.NoError(t, err)

	mi.AssertExpectations(t)
	ml.AssertExpectations(t)
	mo.AssertExpectations(t)

	for n, c := range cs {
		assert.Equalf(t, cfgs[n], c.Config, "config not passed through correctly for %s", n)

		if !ystring.IsBlank(c.SelectedMOpt) {
			checkSelection(t, "MOpt", n, c.SelectedMOpt, dms.Slice(), c.MOpt)
		}
		if c.SelectedOpt != nil {
			checkSelection(t, "Opt", n, c.SelectedOpt.Name, dls.Slice(), c.Opt)
		}
	}
}

func checkSelection(t *testing.T, ty, n, chosen string, defaults []string, sel *optlevel.Selection) {
	t.Helper()

	allowed := defaults
	if sel != nil {
		allowed = append(allowed, sel.Enabled...)
		assert.NotContainsf(t, sel.Disabled, chosen, "selected %s for %s (%s) disabled", ty, n, chosen)
	}
	assert.Containsf(t, allowed, chosen, "selected %s for %s (%s) not allowed", ty, n, chosen)
}

// mockCompilerLister mocks the CompilerLister interface.
type mockCompilerLister struct {
	mock.Mock
}

// ListCompilers mocks the eponymous interface method.
func (m *mockCompilerLister) ListCompilers(ctx context.Context, mid id.ID) (map[string]compiler.Config, error) {
	args := m.Called(ctx, mid)
	return args.Get(0).(map[string]compiler.Config), args.Error(1)
}

// mockInspector mocks the Inspector interface.
type mockInspector struct {
	mock.Mock
}

// DefaultOptLevels mocks the eponymous interface method.
func (m *mockInspector) DefaultOptLevels(c *compiler.Config) (stringhelp.Set, error) {
	args := m.Called(c)
	return args.Get(0).(stringhelp.Set), args.Error(1)
}

// OptLevels mocks the eponymous interface method.
func (m *mockInspector) OptLevels(c *compiler.Config) (map[string]optlevel.Level, error) {
	args := m.Called(c)
	return args.Get(0).(map[string]optlevel.Level), args.Error(1)
}

// DefaultMOpts mocks the eponymous interface method.
func (m *mockInspector) DefaultMOpts(c *compiler.Config) (stringhelp.Set, error) {
	args := m.Called(c)
	return args.Get(0).(stringhelp.Set), args.Error(1)
}

// mockCompilerObserver mocks the CompilerObserver interface.
type mockCompilerObserver struct {
	mock.Mock
}

// OnCompilerPlanStart mocks the eponymous interface method.
func (m *mockCompilerObserver) OnCompilerPlanStart(ncompilers int) {
	_ = m.Called(ncompilers)
}

// OnCompilerPlan mocks the eponymous interface method.
func (m *mockCompilerObserver) OnCompilerPlan(c compiler.Named) {
	_ = m.Called(c)
}

// OnCompilerPlanFinish mocks the eponymous interface method.
func (m *mockCompilerObserver) OnCompilerPlanFinish() {
	_ = m.Called()
}