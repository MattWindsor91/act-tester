// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package lifter_test

import (
	"testing"

	"github.com/c4-project/c4t/internal/id"
	"github.com/c4-project/c4t/internal/stage/lifter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPathset_Prepare tests that Pathset.Prepare makes directories properly.
func TestPathset_Prepare(t *testing.T) {
	td := t.TempDir()
	ps := lifter.NewPathset(td)

	arches := []id.ID{id.ArchArm8, id.ArchPPCPOWER9, id.ArchX8664}
	subjects := []string{"foo", "bar", "baz"}

	// We can't check that the directories don't exist up-front, because the call to Path checks that the directory
	// has been prepared.

	err := ps.Prepare(arches, subjects)
	require.NoError(t, err, "preparing lifter pathset")

	for _, a := range arches {
		for _, s := range subjects {
			d, err := ps.Path(a, s)
			if assert.NoErrorf(t, err, "calculating path for arch=%q subject=%s", a, s) {
				assert.DirExistsf(t, d, "dir must exist for arch=%q subject=%s", a, s)
			}
		}
	}
}
