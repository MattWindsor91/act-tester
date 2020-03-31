// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package director

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"

	"github.com/MattWindsor91/act-tester/internal/director/observer"

	"github.com/MattWindsor91/act-tester/internal/director/pathset"

	"github.com/MattWindsor91/act-tester/internal/model/id"

	"github.com/MattWindsor91/act-tester/internal/transfer/remote"

	"github.com/MattWindsor91/act-tester/internal/director/mach"

	"github.com/MattWindsor91/act-tester/internal/controller/lifter"

	"github.com/MattWindsor91/act-tester/internal/controller/fuzzer"

	"github.com/MattWindsor91/act-tester/internal/controller/planner"
	"github.com/MattWindsor91/act-tester/internal/model/plan"

	"github.com/MattWindsor91/act-tester/internal/config"
	"github.com/MattWindsor91/act-tester/internal/helper/iohelp"
)

// The maximum permitted number of times a loop can error out consecutively before the tester fails.
const maxConsecutiveErrors = 10

// Instance contains the state necessary to run a single machine loop of a director.
type Instance struct {
	// MachConfig contains the machine config for this machine.
	MachConfig config.Machine
	// SSHConfig contains top-level SSH configuration.
	SSHConfig *remote.Config
	// StageConfig is the configuration for this instance's stages.
	StageConfig *StageConfig

	// ID is the ID for this machine.
	ID id.ID

	// InFiles is the list of files to use as the base corpus for this machine loop.
	InFiles []string

	// Env contains the parts of the director's config that tell it how to do various environmental tasks.
	Env *Env

	// Logger points to a logger for this machine's loop.
	Logger *log.Logger

	// Observers is this machine's observer set.
	Observers []observer.Instance

	// SavedPaths contains the save pathset for this machine.
	SavedPaths *pathset.Saved
	// ScratchPaths contains the scratch pathset for this machine.
	ScratchPaths *pathset.Scratch

	// Quantities contains the quantity set for this machine.
	Quantities config.QuantitySet
}

// Run runs this machine's testing loop.
func (i *Instance) Run(ctx context.Context) error {
	i.Logger = iohelp.EnsureLog(i.Logger)
	if err := i.check(); err != nil {
		return err
	}

	i.Logger.Print("preparing scratch directories")
	if err := i.ScratchPaths.Prepare(); err != nil {
		return err
	}

	i.Logger.Print("creating stage configurations")
	sc, err := i.makeStageConfig()
	if err != nil {
		return err
	}
	i.Logger.Print("checking stage configurations")
	if err := sc.Check(); err != nil {
		return err
	}

	i.Logger.Print("starting loop")
	return i.mainLoop(ctx, sc)
}

// check makes sure this machine has a valid configuration before starting loops.
func (i *Instance) check() error {
	if i.ScratchPaths == nil {
		return fmt.Errorf("%w: paths for machine %s", iohelp.ErrPathsetNil, i.ID.String())
	}

	if i.Env == nil {
		return errors.New("no environment configuration")
	}

	// TODO(@MattWindsor): check SSHConfig?

	return nil
}

// mainLoop performs the main testing loop for one machine.
func (i *Instance) mainLoop(ctx context.Context, sc *StageConfig) error {
	var (
		iter    uint64
		nErrors uint
	)
	for {
		if err := i.pass(ctx, iter, sc); err != nil {
			// This serves to stop the tester if we get stuck in a rapid failure loop on a particular machine.
			// TODO(@MattWindsor91): ideally this should be timing the gap between errors, so that we stop if there
			// are too many errors happening too quickly.
			nErrors++
			if maxConsecutiveErrors < nErrors {
				return fmt.Errorf("too many consecutive errors; last error was: %w", err)
			}
			i.Logger.Println("ERROR:", err)
		} else {
			nErrors = 0
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		iter++
	}
}

// pass performs one iteration of the main testing loop (number iter) for one machine.
func (i *Instance) pass(ctx context.Context, iter uint64, sc *StageConfig) error {
	var (
		p   *plan.Plan
		err error
	)

	observer.OnIteration(iter, time.Now(), i.Observers...)

	for _, s := range Stages {
		if p, err = s.Run(sc, ctx, p); err != nil {
			return fmt.Errorf("in %s stage: %w", s.Name, err)
		}
		if err = i.dump(s.Name, p); err != nil {
			return fmt.Errorf("when dumping after %s stage: %w", s.Name, err)
		}
	}

	return nil
}

func (i *Instance) makeStageConfig() (*StageConfig, error) {
	obs := observer.LowerToBuilder(i.Observers)

	p, err := i.makePlanner(obs)
	if err != nil {
		return nil, fmt.Errorf("when making planner: %w", err)
	}
	f, err := i.makeFuzzerConfig(obs)
	if err != nil {
		return nil, fmt.Errorf("when making fuzzer config: %w", err)
	}
	l, err := i.makeLifterConfig(obs)
	if err != nil {
		return nil, fmt.Errorf("when making lifter config: %w", err)
	}
	m, err := mach.New(i.Observers, i.ScratchPaths.DirRun, i.SSHConfig, i.MachConfig.SSH)
	if err != nil {
		return nil, fmt.Errorf("when making machine-exec config: %w", err)
	}
	sc := StageConfig{
		InFiles: i.InFiles,
		Plan:    p,
		Fuzz:    f,
		Lift:    l,
		Mach:    m,
		Save:    i.makeSave(),
	}
	return &sc, nil
}

func (i *Instance) makeSave() *Save {
	return &Save{
		Logger:    i.Logger,
		Observers: i.Observers,
		NWorkers:  10, // TODO(@MattWindsor91): get this from somewhere
		Paths:     i.SavedPaths,
	}
}

func (i *Instance) makePlanner(obs []builder.Observer) (*planner.Planner, error) {
	p := planner.Planner{
		Source:    i.Env.Planner,
		Logger:    i.Logger,
		Observers: obs,
		MachineID: i.ID,
	}
	return &p, nil
}

func (i *Instance) makeFuzzerConfig(obs []builder.Observer) (*fuzzer.Config, error) {
	fz := i.Env.Fuzzer
	if fz == nil {
		return nil, errors.New("no single fuzzer provided")
	}

	fc := fuzzer.Config{
		Driver:     fz,
		Logger:     i.Logger,
		Observers:  obs,
		Paths:      fuzzer.NewPathset(i.ScratchPaths.DirFuzz),
		Quantities: i.Quantities.Fuzz,
	}

	return &fc, nil
}

func (i *Instance) makeLifterConfig(obs []builder.Observer) (*lifter.Config, error) {
	hm := i.Env.Lifter
	if hm == nil {
		return nil, errors.New("no single fuzzer provided")
	}

	lc := lifter.Config{
		Maker:     hm,
		Logger:    i.Logger,
		Observers: obs,
		Paths:     lifter.NewPathset(i.ScratchPaths.DirLift),
	}

	return &lc, nil
}

// dump dumps a plan p to its expected plan file given the stage name name.
func (i *Instance) dump(name string, p *plan.Plan) error {
	fname := i.ScratchPaths.PlanForStage(name)
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("while opening plan file for %s: %w", name, err)
	}
	if err := p.Dump(f); err != nil {
		_ = f.Close()
		return fmt.Errorf("while writing plan file for %s: %w", name, err)
	}
	return f.Close()
}