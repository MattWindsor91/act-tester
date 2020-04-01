// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package dash

import (
	"fmt"

	"github.com/MattWindsor91/act-tester/internal/model/id"
	"github.com/MattWindsor91/act-tester/internal/model/subject"

	"github.com/MattWindsor91/act-tester/internal/helper/iohelp"
	"github.com/mum4k/termdash/cell"

	"github.com/MattWindsor91/act-tester/internal/model/corpus/builder"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/text"
)

// actionObserver is the portion of the dashboard that observes action builds.
type actionObserver struct {
	log   *text.Text
	gauge *gauge.Gauge

	nreqs, ndone int
}

func NewCorpusObserver() (*actionObserver, error) {
	var (
		c   actionObserver
		err error
	)

	if c.gauge, err = gauge.New(); err != nil {
		return nil, err
	}

	if c.log, err = text.New(text.RollContent()); err != nil {
		return nil, err
	}

	return &c, nil
}

func (o *actionObserver) gridRows() []grid.Element {
	return []grid.Element{
		grid.RowHeightFixed(1, grid.Widget(o.gauge)),
		grid.RowHeightFixed(1, grid.Widget(o.log)),
	}
}

// OnBuildStart sets up an observer for a test phase with manifest m.
func (o *actionObserver) OnBuildStart(m builder.Manifest) {
	o.onTaskStart(m.Name, m.NReqs)
}

// OnBuildRequest acknowledges a action-builder request.
func (o *actionObserver) OnBuildRequest(r builder.Request) {
	switch {
	case r.Add != nil:
		o.onAdd(r.Name)
	case r.Compile != nil:
		o.onCompile(r.Name, r.Compile)
	case r.Harness != nil:
		o.onHarness(r.Name, r.Harness)
	case r.Run != nil:
		o.onRun(r.Name, r.Run)
	}
}

// onAdd acknowledges the addition of a subject to a action being built.
func (o *actionObserver) onAdd(sname string) {
	o.logAndStepGauge("ADD", sname, colorAdd)
}

// onCompile acknowledges the addition of a compilation to a action being built.
func (o *actionObserver) onCompile(sname string, b *builder.Compile) {
	c := colorCompile
	desc := idQualSubjectDesc(sname, b.CompilerID)

	if !b.Result.Success {
		c = colorFailed
		desc += " [FAILED]"
	}

	o.logAndStepGauge("COMPILE", desc, c)
}

// onHarness acknowledges the addition of a harness to a action being built.
func (o *actionObserver) onHarness(sname string, b *builder.Harness) {
	o.logAndStepGauge("LIFT", idQualSubjectDesc(sname, b.Arch), colorHarness)
}

// onRun acknowledges the addition of a run to a action being built.
func (o *actionObserver) onRun(sname string, b *builder.Run) {
	desc := idQualSubjectDesc(sname, b.CompilerID)
	suff, c := runSuffixAndColour(b.Result.Status)
	o.logAndStepGauge("RUN", desc+suff, c)
}

func runSuffixAndColour(s subject.Status) (string, cell.Color) {
	switch s {
	case subject.StatusFlagged:
		return " [FLAGGED]", colorFlagged
	case subject.StatusRunTimeout:
		return " [TIMEOUT]", colorTimeout
	case subject.StatusCompileFail:
		return " [FAILED]", colorFailed
	default:
		return "", colorRun
	}
}

// OnBuildFinish acknowledges the end of a run.
func (o *actionObserver) OnBuildFinish() {
	_ = o.log.Write("-- DONE --\n")
}

// OnCopyStart acknowledges the start of a file copy.
func (o *actionObserver) OnCopyStart(nfiles int) {
	o.onTaskStart("COPYING FILES", nfiles)
}

// OnCopy acknowledges one step of a file copy.
func (o *actionObserver) OnCopy(dst, src string) {
	desc := fmt.Sprintf("%s -> %s", src, dst)
	o.logAndStepGauge("COPY", desc, colorCopy)
}

// OnCopyFinish acknowledges the end of a file copy.
func (o *actionObserver) OnCopyFinish() {
	// TODO(@MattWindsor91): abstract this properly
	o.OnBuildFinish()
}

func (o *actionObserver) onTaskStart(name string, n int) {
	_ = o.log.Write(fmt.Sprintf("-- %s --\n", name))

	o.nreqs = n
	o.ndone = 0
	_ = o.gauge.Absolute(o.ndone, o.nreqs)
}

func (o *actionObserver) logError(err error) {
	if err == nil {
		return
	}
	_ = o.log.Write(err.Error(), text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
}

func (o *actionObserver) reset() {
	o.log.Reset()
}

// logAndStepGauge logs a request with name rq and summary desc, then repopulates the gauge.
// It uses c as the colour for both.
func (o *actionObserver) logAndStepGauge(rq, desc string, c cell.Color) {
	lerr := o.logStep(rq, desc, c)
	serr := o.stepGauge(c)
	o.logError(iohelp.FirstError(lerr, serr))
}

// logStep logs an observed builder request with name rq and summary desc to the per-machine log.
// It colours the log with c.
func (o *actionObserver) logStep(rq, desc string, c cell.Color) error {
	ferr := o.log.Write(rq, text.WriteCellOpts(cell.FgColor(c)))
	lerr := o.log.Write(" " + desc + "\n")
	return iohelp.FirstError(ferr, lerr)
}

// stepGauge increments the gauge and sets its colour to c.
func (o *actionObserver) stepGauge(c cell.Color) error {
	o.ndone++
	return o.gauge.Absolute(o.ndone, o.nreqs, gauge.Color(c))
}

func idQualSubjectDesc(sname string, id id.ID) string {
	return fmt.Sprintf("%s (@%s)", sname, id)
}
