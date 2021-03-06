// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package mkdoc

import (
	"os"
	"path/filepath"

	c "github.com/urfave/cli/v2"
)

const extMan = ".8"
const fileMarkdown = "README.md"

type method struct {
	name string
	make func() (string, error)
}

func methodsOf(app *c.App) map[string]method {
	return map[string]method{
		"manpage":  {name: app.Name + extMan, make: app.ToMan},
		"markdown": {name: fileMarkdown, make: app.ToMarkdown},
	}
}

func (m method) run(outdir string) error {
	s, err := m.make()
	if err != nil {
		return err
	}
	fname := filepath.Join(outdir, m.name)
	return os.WriteFile(fname, []byte(s), 0744)
}
