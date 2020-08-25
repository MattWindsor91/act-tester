// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package act_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MattWindsor91/act-tester/internal/act"
	"github.com/MattWindsor91/act-tester/internal/machine"
	"github.com/MattWindsor91/act-tester/internal/model/job"
)

// ExampleWriteFuzzConf is a testable example for WriteFuzzConf.
func ExampleWriteFuzzConf() {
	noMachine := job.Fuzzer{}
	if err := act.WriteFuzzConf(os.Stdout, noMachine); err != nil {
		fmt.Println("ERROR:", err)
	}

	fmt.Println("")

	mach := job.Fuzzer{Machine: &machine.Machine{Cores: 4}}
	if err := act.WriteFuzzConf(os.Stdout, mach); err != nil {
		fmt.Println("ERROR:", err)
	}

	// Output:
	// # AUTOGENERATED BY TESTER
	// fuzz {
	// }
	//
	// # AUTOGENERATED BY TESTER
	// fuzz {
	// ## MACHINE SPECIFIC OVERRIDES ##
	//   # Set to number of cores in machine to prevent thrashing.
	//   set param cap.threads to 4
	// }
}

// TestMakeFuzzConfFile tests to make sure that MakeFuzzConfFile makes a valid file containing the same thing as
// using WriteFuzzConf.
func TestMakeFuzzConfFile(t *testing.T) {
	t.Parallel()

	cases := map[string]job.Fuzzer{
		"no-machine": {Machine: nil},
		"machine":    {Machine: &machine.Machine{Cores: 4}},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, act.WriteFuzzConf(&buf, c), "writing config to buffer shouldn't error")
			want := buf.String()

			cf, cerr := act.MakeFuzzConfFile(c)
			require.NoError(t, cerr, "saving config to file shouldn't error")
			require.FileExists(t, cf, "config file should exist")
			defer func() { _ = os.Remove(cf) }()

			got, rerr := ioutil.ReadFile(cf)
			require.NoError(t, rerr, "loading config from file shouldn't error")

			require.Equal(t, want, string(got), "config didn't match")
		})
	}
}
