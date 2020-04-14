// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package director

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/1set/gut/yos"

	"github.com/MattWindsor91/act-tester/internal/director/observer"

	"github.com/MattWindsor91/act-tester/internal/model/normalise"

	"github.com/MattWindsor91/act-tester/internal/director/pathset"
	"github.com/MattWindsor91/act-tester/internal/helper/iohelp"
	"github.com/MattWindsor91/act-tester/internal/model/corpus"
	"github.com/MattWindsor91/act-tester/internal/model/corpus/collate"
	"github.com/MattWindsor91/act-tester/internal/model/plan"
	"github.com/MattWindsor91/act-tester/internal/model/subject"
)

// Save contains the state used when saving 'interesting' subjects.
type Save struct {
	// Logger is the logger to use when announcing that we're saving subjects.
	Logger *log.Logger

	// Observers is the list of instance observers.
	Observers []observer.Instance

	// NWorkers is the number of workers to use for the collator.
	NWorkers int

	// Paths contains the pathset used to save subjects for a particular machine.
	Paths *pathset.Saved
}

// Run runs the saving stage over plan p.
// It returns p unchanged; this is for signature compatibility with the other director stages.
func (s *Save) Run(ctx context.Context, p *plan.Plan) (*plan.Plan, error) {
	if s.Paths == nil {
		return nil, iohelp.ErrPathsetNil
	}

	s.Logger = iohelp.EnsureLog(s.Logger)

	if err := s.Paths.Prepare(); err != nil {
		return nil, err
	}

	coll, err := collate.Collate(ctx, p.Corpus, s.NWorkers)
	if err != nil {
		return nil, fmt.Errorf("when collating: %w", err)
	}
	observer.OnCollation(coll, s.Observers...)

	for st, c := range coll.ByStatus() {
		if st < subject.FirstBadStatus || len(c) == 0 {
			continue
		}
		if err := s.saveBucket(st, c, p, p.Header.Creation); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (s *Save) saveBucket(st subject.Status, c corpus.Corpus, p *plan.Plan, creation time.Time) error {
	if err := s.prepareDir(st, creation); err != nil {
		return err
	}
	if err := s.writePlan(st, p, creation); err != nil {
		return err
	}
	return s.tarSubjects(st, c, creation)
}

func (s *Save) prepareDir(st subject.Status, creation time.Time) error {
	dir, err := s.Paths.SubjectDir(st, creation)
	if err != nil {
		return err
	}
	return yos.MakeDir(dir)
}

func (s *Save) writePlan(st subject.Status, p *plan.Plan, creation time.Time) error {
	path, err := s.Paths.PlanFile(st, creation)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = p.Dump(f)
	cerr := f.Close()
	return iohelp.FirstError(err, cerr)
}

func (s *Save) tarSubjects(st subject.Status, corp corpus.Corpus, creation time.Time) error {
	for name, sub := range corp {
		if err := s.tarSubject(st, name, sub, creation); err != nil {
			return err
		}
	}
	return nil
}

func (s *Save) tarSubject(st subject.Status, name string, sub subject.Subject, creation time.Time) error {
	tarpath, err := s.Paths.SubjectTarFile(name, st, creation)
	if err != nil {
		return err
	}
	s.Logger.Printf("archiving %s (to %q)", name, tarpath)
	if err := s.tarSubjectToPath(sub, tarpath); err != nil {
		return fmt.Errorf("tarring subject %s: %w", name, err)
	}
	return nil
}

func (s *Save) tarSubjectToPath(sub subject.Subject, tarpath string) error {
	tarfile, err := os.Create(tarpath)
	if err != nil {
		return fmt.Errorf("create %s: %w", tarpath, err)
	}
	gzw := gzip.NewWriter(tarfile)
	tarw := tar.NewWriter(gzw)

	werr := s.tarSubjectToWriter(sub, tarw)
	terr := tarw.Close()
	gerr := gzw.Close()

	if werr != nil {
		return werr
	}
	if terr != nil {
		return fmt.Errorf("closing tar: %w", terr)
	}
	if gerr != nil {
		return fmt.Errorf("closing gzip: %w", gerr)
	}
	return nil
}

func (s *Save) tarSubjectToWriter(sub subject.Subject, tarw *tar.Writer) error {
	fs, err := filesToTar(sub)
	if err != nil {
		return err
	}
	for wpath, norm := range fs {
		rpath := norm.Original
		if err := s.tarFileToWriter(rpath, wpath, tarw); err != nil {
			return fmt.Errorf("archiving %q: %w", rpath, err)
		}
	}
	return nil
}

func filesToTar(s subject.Subject) (map[string]normalise.Normalisation, error) {
	n := normalise.NewNormaliser("")
	if _, err := n.Subject(s); err != nil {
		return nil, err
	}
	return n.Mappings, nil
}

// tarFileToWriter tars the file at rpath to wpath within the tar archive represented by tarw.
// If rpath is empty, no tarring occurs.
func (s *Save) tarFileToWriter(rpath, wpath string, tarw *tar.Writer) error {
	if rpath == "" {
		return nil
	}

	hdr, err := tarFileHeader(rpath, wpath)
	if err != nil {
		return s.rescueNotExistError(err, rpath)
	}
	if err := tarw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	f, err := os.Open(rpath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", rpath, err)
	}
	_, err = io.Copy(tarw, f)
	cerr := f.Close()
	if err != nil {
		return fmt.Errorf("archiving %s: %w", rpath, err)
	}
	return cerr
}

func (s *Save) rescueNotExistError(err error, rpath string) error {
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.Logger.Println("file missing when archiving error:", rpath)
	return nil
}

func tarFileHeader(rpath, wpath string) (*tar.Header, error) {
	info, err := os.Stat(rpath)
	if err != nil {
		return nil, fmt.Errorf("can't stat %s: %w", rpath, err)
	}
	hdr, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil, fmt.Errorf("can't get header for %s: %w", rpath, err)
	}
	hdr.Name = wpath
	return hdr, nil
}
