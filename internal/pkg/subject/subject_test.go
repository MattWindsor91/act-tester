// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package subject_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/MattWindsor91/act-tester/internal/pkg/subject"

	"github.com/MattWindsor91/act-tester/internal/pkg/testhelp"

	"github.com/MattWindsor91/act-tester/internal/pkg/model"
)

// ExampleSubject_BestLitmus is a testable example for BestLitmus.
func ExampleSubject_BestLitmus() {
	s1 := subject.Subject{Litmus: "foo.litmus"}
	b1, _ := s1.BestLitmus()

	// This subject has a fuzzed litmus file, which takes priority.
	s2 := subject.Subject{Litmus: "foo.litmus", Fuzz: &subject.FuzzFileset{Litmus: "bar.litmus"}}
	b2, _ := s2.BestLitmus()

	fmt.Println("s1:", b1)
	fmt.Println("s2:", b2)

	// Output:
	// s1: foo.litmus
	// s2: bar.litmus
}

// ExampleSubject_CompileResult is a testable example for CompileResult.
func ExampleSubject_CompileResult() {
	s := subject.Subject{Compiles: map[string]subject.CompileResult{
		"localhost:gcc":   {Success: true, Files: subject.CompileFileset{Bin: "a.out", Log: "gcc.log"}},
		"spikemuth:clang": {Success: false, Files: subject.CompileFileset{Bin: "a.out", Log: "clang.log"}},
	}}
	lps, _ := s.CompileResult(model.MachQualID{MachineID: model.IDFromString("localhost"), ID: model.IDFromString("gcc")})
	sps, _ := s.CompileResult(model.MachQualID{MachineID: model.IDFromString("spikemuth"), ID: model.IDFromString("clang")})

	fmt.Println("localhost:", lps.Success, lps.Files.Bin, lps.Files.Log)
	fmt.Println("spikemuth:", sps.Success, sps.Files.Bin, sps.Files.Log)

	// Output:
	// localhost: true a.out gcc.log
	// spikemuth: false a.out clang.log
}

// ExampleSubject_Harness is a testable example for Harness.
func ExampleSubject_Harness() {
	s := subject.Subject{Harnesses: map[string]subject.Harness{
		"localhost:x86.64": {Dir: "foo", Files: []string{"bar", "baz"}},
		"spikemuth:arm":    {Dir: "foobar", Files: []string{"barbaz"}},
	}}
	lps, _ := s.Harness(model.MachQualID{MachineID: model.IDFromString("localhost"), ID: model.ArchX8664})
	sps, _ := s.Harness(model.MachQualID{MachineID: model.IDFromString("spikemuth"), ID: model.ArchArm})

	for _, l := range lps.Files {
		fmt.Println(l)
	}
	for _, s := range sps.Files {
		fmt.Println(s)
	}

	// Output:
	// bar
	// baz
	// barbaz
}

// TestSubject_CompileResult_Missing checks that trying to get a harness path for a missing machine/emits pair triggers
// the appropriate error.
func TestSubject_CompileResult_Missing(t *testing.T) {
	var s subject.Subject
	_, err := s.CompileResult(model.MachQualID{
		MachineID: model.IDFromString("localhost"),
		ID:        model.IDFromString("gcc"),
	})
	testhelp.ExpectErrorIs(t, err, subject.ErrMissingCompile, "missing compile result path")
}

// TestSubject_AddCompileResult checks that AddCompileResult is working properly.
func TestSubject_AddCompileResult(t *testing.T) {
	var s subject.Subject
	c := subject.CompileResult{
		Success: true,
		Files: subject.CompileFileset{
			Bin: "a.out",
			Log: "gcc.log",
		},
	}

	mcomp := model.MachQualID{
		MachineID: model.IDFromString("localhost"),
		ID:        model.IDFromString("gcc"),
	}

	t.Run("initial-add", func(t *testing.T) {
		if err := s.AddCompileResult(mcomp, c); err != nil {
			t.Fatalf("err when adding compile to empty subject: %v", err)
		}
	})
	t.Run("add-get", func(t *testing.T) {
		c2, err := s.CompileResult(mcomp)
		if err != nil {
			t.Fatalf("err when getting added compile: %v", err)
		}
		if !reflect.DeepEqual(c2, c) {
			t.Fatalf("added compile (%v) came back wrong (%v)", c2, c)
		}
	})
	t.Run("add-dupe", func(t *testing.T) {
		err := s.AddCompileResult(mcomp, subject.CompileResult{})
		testhelp.ExpectErrorIs(t, err, subject.ErrDuplicateCompile, "adding compile twice")
	})
}

// TestSubject_Harness_Missing checks that trying to get a harness path for a missing machine/emits pair triggers
// the appropriate error.
func TestSubject_Harness_Missing(t *testing.T) {
	var s subject.Subject
	_, err := s.Harness(model.MachQualID{MachineID: model.IDFromString("localhost"), ID: model.IDFromString("x86.64")})
	testhelp.ExpectErrorIs(t, err, subject.ErrMissingHarness, "missing harness path")
}

// TestSubject_AddHarness checks that AddHarness is working properly.
func TestSubject_AddHarness(t *testing.T) {
	var s subject.Subject
	h := subject.Harness{
		Dir:   "foo",
		Files: []string{"bar", "baz"},
	}

	march := model.MachQualID{
		MachineID: model.IDFromString("localhost"),
		ID:        model.ArchX8664,
	}

	t.Run("initial-add", func(t *testing.T) {
		if err := s.AddHarness(march, h); err != nil {
			t.Fatalf("err when adding harness to empty subject: %v", err)
		}
	})
	t.Run("add-get", func(t *testing.T) {
		h2, err := s.Harness(march)
		if err != nil {
			t.Fatalf("err when getting added harness: %v", err)
		}
		if !reflect.DeepEqual(h2, h) {
			t.Fatalf("added harness (%v) came back wrong (%v)", h2, h)
		}
	})
	t.Run("add-dupe", func(t *testing.T) {
		err := s.AddHarness(march, subject.Harness{})
		testhelp.ExpectErrorIs(t, err, subject.ErrDuplicateHarness, "adding harness twice")
	})
}

// TestSubject_BestLitmus tests a few cases of BestLitmus.
// It should be more comprehensive than the examples.
func TestSubject_BestLitmus(t *testing.T) {
	// Note that the presence of 'err' overrides 'want'.
	cases := map[string]struct {
		s    subject.Subject
		err  error
		want string
	}{
		"zero":             {s: subject.Subject{}, err: subject.ErrNoBestLitmus, want: ""},
		"zero-fuzz":        {s: subject.Subject{Fuzz: &subject.FuzzFileset{}}, err: subject.ErrNoBestLitmus, want: ""},
		"litmus-only":      {s: subject.Subject{Litmus: "foo"}, err: nil, want: "foo"},
		"litmus-only-fuzz": {s: subject.Subject{Litmus: "foo", Fuzz: &subject.FuzzFileset{}}, err: nil, want: "foo"},
		"fuzz":             {s: subject.Subject{Litmus: "foo", Fuzz: &subject.FuzzFileset{Litmus: "bar"}}, err: nil, want: "bar"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := c.s.BestLitmus()
			switch {
			case err != nil && c.err == nil:
				t.Errorf("unexpected BestLitmus(%v) error: %v", c.s, err)
			case err != nil && !errors.Is(err, c.err):
				t.Errorf("wrong BestLitmus(%v) error: got %v; want %v", c.s, err, c.err)
			case err == nil && c.err != nil:
				t.Errorf("no BestLitmus(%v) error; want %v", c.s, err)
			case err == nil && got != c.want:
				t.Errorf("BestLitmus(%v)=%q; want %q", c.s, got, c.want)
			}
		})
	}
}