// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package corpus

import (
	"github.com/c4-project/c4t/internal/helper/stringhelp"
)

// Merge merges corpora into a single corpus.
// If corpora is empty or nil, it returns nil.
// If there is only one corpus in corpora, it just deep-copies that corpus.
// Otherwise, it produces a new corpus with each subject's name prefixed by its corpus's name in corpora.
func Merge(corpora map[string]Corpus) (Corpus, error) {
	if len(corpora) == 0 {
		return nil, nil
	}
	if len(corpora) == 1 {
		for _, c := range corpora {
			return c.Copy(), nil
		}
	}
	return actuallyMerge(corpora)
}

func actuallyMerge(corpora map[string]Corpus) (Corpus, error) {
	result := make(Corpus)
	for cname, c := range corpora {
		for sname, s := range c {
			if err := result.Add(*s.AddName(MergedName(cname, sname))); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// MergedName is the name that sname will appear under in a merged corpus where the original corpus name was cname.
func MergedName(cname, sname string) string {
	return stringhelp.JoinNonEmpty("/", cname, sname)
}
