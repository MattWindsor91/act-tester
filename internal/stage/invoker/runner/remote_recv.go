// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package runner

import (
	"context"
	"fmt"
	"path"

	copy2 "github.com/MattWindsor91/act-tester/internal/copier"

	"github.com/MattWindsor91/act-tester/internal/model/filekind"

	"github.com/MattWindsor91/act-tester/internal/model/corpus"
	"github.com/MattWindsor91/act-tester/internal/model/normaliser"
	"github.com/MattWindsor91/act-tester/internal/model/subject"
	"github.com/MattWindsor91/act-tester/internal/plan"
)

// Recv copies bits of remp into locp, including run information and any compiler failures.
// It uses SFTP to transfer back any compile logs.
func (r *RemoteRunner) Recv(ctx context.Context, locp, remp *plan.Plan) (*plan.Plan, error) {
	locp.Metadata.Stages = remp.Metadata.Stages

	err := locp.Corpus.Map(func(sn *subject.Named) error {
		return r.recvSubject(ctx, sn, remp.Corpus)
	})
	return locp, err
}

func (r *RemoteRunner) recvSubject(ctx context.Context, ls *subject.Named, rcorp corpus.Corpus) error {
	norm := normaliser.New(path.Join(r.localRoot, ls.Name))
	rs, ok := rcorp[ls.Name]
	if !ok {
		return fmt.Errorf("subject not in remote corpus: %s", ls.Name)
	}
	ns, err := norm.Normalise(rs)
	if err != nil {
		return fmt.Errorf("can't normalise subject: %w", err)
	}
	ls.Runs = ns.Runs
	ls.Compiles = ns.Compiles
	return r.recvMapping(ctx, norm.Mappings.RenamesMatching(filekind.Any, filekind.InCompile))
}

func (r *RemoteRunner) recvMapping(ctx context.Context, ms map[string]string) error {
	cli, err := r.runner.NewSFTP()
	if err != nil {
		return err
	}

	perr := copy2.RecvMapping(ctx, (*copy2.SFTP)(cli), ms, r.observers...)
	cerr := cli.Close()

	if perr != nil {
		return perr
	}
	return cerr
}
