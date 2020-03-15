// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package director

import (
	"context"
	"errors"
	"fmt"

	"github.com/MattWindsor91/act-tester/internal/pkg/corpus"

	"github.com/MattWindsor91/act-tester/internal/pkg/fuzzer"
	"github.com/MattWindsor91/act-tester/internal/pkg/lifter"
	"github.com/MattWindsor91/act-tester/internal/pkg/plan"
	"github.com/MattWindsor91/act-tester/internal/pkg/planner"
)

// StageConfig groups together the stage configuration for of a director instance.
type StageConfig struct {
	// InFiles contains the input files for the instance.
	InFiles []string
	// Plan contains configuration for the instance's plan stage.
	Plan *planner.Planner
	// Fuzz contains configuration for the instance's fuzz stage.
	Fuzz *fuzzer.Config
	// Lift contains configuration for the instance's lift stage.
	Lift *lifter.Config
}

var ErrStageConfigMissing = errors.New("stage config missing")

// Check makes sure the StageConfig has all configuration elements present.
func (c *StageConfig) Check() error {
	if len(c.InFiles) == 0 {
		return fmt.Errorf("%w: no input files", corpus.ErrNone)
	}
	if c.Plan == nil {
		return fmt.Errorf("%w: %s", ErrStageConfigMissing, StagePlan)
	}
	if c.Fuzz == nil {
		return fmt.Errorf("%w: %s", ErrStageConfigMissing, StageFuzz)
	}
	if c.Lift == nil {
		return fmt.Errorf("%w: %s", ErrStageConfigMissing, StageLift)
	}
	return nil
}

// Stage is the type of encapsulated director stages.
type stage struct {
	// Name is the name of the stage, which appears in logging and errors.
	Name string
	// Run is the function to use to run the stage.
	Run func(*StageConfig, context.Context, *plan.Plan) (*plan.Plan, error)
}

const (
	StagePlan = "plan"
	StageFuzz = "fuzz"
	StageLift = "lift"
)

// Stages is the list of director stages.
var Stages = []stage{
	{
		Name: StagePlan,
		Run: func(c *StageConfig, ctx context.Context, _ *plan.Plan) (*plan.Plan, error) {
			return c.Plan.Plan(ctx, c.InFiles)
		},
	},
	{
		Name: StageFuzz,
		Run: func(c *StageConfig, ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
			return c.Fuzz.Run(ctx, p)
		},
	},
	{
		Name: StageLift,
		Run: func(c *StageConfig, ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
			return c.Lift.Run(ctx, p)
		},
	},
}