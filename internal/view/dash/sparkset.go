// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package dash

import (
	"github.com/MattWindsor91/act-tester/internal/model/corpus/collate"
	"github.com/MattWindsor91/act-tester/internal/model/subject"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/widgets/sparkline"
)

// sparkset contains the sparklines for a machine.
type sparkset struct {
	// statusLines contains one sparkline for each status.
	// (This _includes_ StatusUnknown to simplify calculations later, but we don't display it as a line.)
	statusLines [subject.NumStatus]*sparkline.SparkLine
}

func newSparkset() (*sparkset, error) {
	var (
		s   sparkset
		err error
	)

	for i := subject.StatusOk; i < subject.NumStatus; i++ {
		if s.statusLines[i], err = sparkline.New(
			sparkline.Color(statusColours[i]), sparkline.Label(i.String()),
		); err != nil {
			return nil, err
		}
	}
	return &s, err
}

func (s *sparkset) sparkLines() []*sparkline.SparkLine {
	return s.statusLines[subject.StatusOk:]
}

func (s *sparkset) grid() []grid.Element {
	sls := s.sparkLines()
	els := make([]grid.Element, len(sls))
	for i, sl := range sls {
		els[i] = grid.RowHeightFixed(2, grid.Widget(sl))
	}
	return els
}

func (s *sparkset) sparkCollation(c *collate.Collation) error {
	for i := subject.StatusOk; i < subject.NumStatus; i++ {
		if err := s.statusLines[i].Add([]int{len(c.ByStatus[i])}); err != nil {
			return err
		}
	}
	return nil
}
