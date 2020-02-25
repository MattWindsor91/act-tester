// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package interop

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrStatsetParse occurs when there is a parse error reading a statset.
var ErrStatsetParse = errors.New("statistic parse error")

// StatDumper is the interface of things that can dump statistics for a litmus test.
type StatDumper interface {
	// DumpStats populates s with statistics gleaned from the Litmus file at path.
	DumpStats(ctx context.Context, s *Statset, path string) error
}

// Statset contains a set of statistics acquired from `act-c dump-stats`.
type Statset struct {
	// Threads is the number of threads.
	Threads int

	// Returns is the number of return statements.
	Returns int

	// LiteralBools is the number of Boolean literals (true, false, etc).
	LiteralBools int
}

// Parse parses a statistics set from r into this statistics set.
// Each statistic should be in the form "name value\n".
func (s *Statset) Parse(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if err := s.parseLine(sc.Text()); err != nil {
			return err
		}
	}
	return sc.Err()
}

func (s *Statset) parseLine(l string) error {
	fs := strings.Fields(l)
	if len(fs) != 2 {
		return fmt.Errorf("%w: %d fields in %q; want 2", ErrStatsetParse, len(fs), fs)
	}

	name := fs[0]
	val, err := strconv.Atoi(fs[1])
	if err != nil {
		return err
	}
	return s.setByName(name, val)
}

func (s *Statset) setByName(name string, val int) error {
	switch name {
	case "threads":
		s.Threads = val
	case "returns":
		s.Returns = val
	case "literal-bools":
		s.LiteralBools = val
	default:
		return fmt.Errorf("%w: unknown stat %s", ErrStatsetParse, name)
	}
	return nil
}