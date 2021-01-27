// Copyright (c) 2020-2021 C4 Project
//
// This file is part of c4t.
// Licenced under the MIT licence; see `LICENSE`.

package c4f_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/c4-project/c4t/internal/model/service/fuzzer"

	"github.com/stretchr/testify/require"

	"github.com/c4-project/c4t/internal/c4f"
	"github.com/c4-project/c4t/internal/machine"
)

// ExampleWriteFuzzConf_empty is a testable example for WriteFuzzConf, showing an empty config.
func ExampleWriteFuzzConf_empty() {
	if err := c4f.WriteFuzzConf(os.Stdout, (fuzzer.Job{})); err != nil {
		fmt.Println("ERROR:", err)
	}

	// Output:
	// # AUTOGENERATED BY TESTER
	// fuzz {
	// }
}

// ExampleWriteFuzzConf_machine is a testable example for WriteFuzzConf, showing the effect of adding a machine.
func ExampleWriteFuzzConf_machine() {
	f := fuzzer.Job{Machine: &machine.Machine{Cores: 4}}
	if err := c4f.WriteFuzzConf(os.Stdout, f); err != nil {
		fmt.Println("ERROR:", err)
	}

	// Output:
	// # AUTOGENERATED BY TESTER
	// fuzz {
	// ## MACHINE SPECIFIC OVERRIDES ##
	//   # Set to number of cores in machine to prevent thrashing.
	//   set param cap.threads to 4
	// }
}

// ExampleWriteFuzzConf_params is a testable example for WriteFuzzConf, showing the effect of adding parameters.
func ExampleWriteFuzzConf_params() {
	// These will be output in ascending alphabetical order of keys.
	f := fuzzer.Job{Config: &fuzzer.Config{Params: map[string]string{
		"int.action.cap.upper":          "1000",
		"int.this.will.not.parse":       "six",
		"bool.mem.unsafe-weaken-orders": "true",
		"bool.action.enable":            "2:1",
		"bool.action.pick-extra":        "false",
		"bool.this.will.not.parse":      ":",
		"action.var.make":               "10",
		"action.var.nope":               "ten",
		"nonsuch":                       "thing",
		"":                              "nope",
	}}}
	if err := c4f.WriteFuzzConf(os.Stdout, f); err != nil {
		fmt.Println("ERROR:", err)
	}

	// Output:
	// # AUTOGENERATED BY TESTER
	// fuzz {
	// ## CONFIGURATION OVERRIDES ##
	//   # unsupported param "": "nope"
	//   action var.make weight 10
	//   # unsupported param "action.var.nope": "ten"
	//   set flag action.enable to ratio 2:1
	//   set flag action.pick-extra to false
	//   set flag mem.unsafe-weaken-orders to true
	//   # unsupported param "bool.this.will.not.parse": ":"
	//   set param action.cap.upper to 1000
	//   # unsupported param "int.this.will.not.parse": "six"
	//   # unsupported param "nonsuch": "thing"
	// }
}

// TestMakeFuzzConfFile tests to make sure that MakeFuzzConfFile makes a valid file containing the same thing as
// using WriteFuzzConf.
func TestMakeFuzzConfFile(t *testing.T) {
	t.Parallel()

	cases := map[string]fuzzer.Job{
		"no-machine": {Machine: nil},
		"machine":    {Machine: &machine.Machine{Cores: 4}},
	}
	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, c4f.WriteFuzzConf(&buf, c), "writing config to buffer shouldn't error")
			want := buf.String()

			cf, cerr := c4f.MakeFuzzConfFile(c)
			require.NoError(t, cerr, "saving config to file shouldn't error")
			require.FileExists(t, cf, "config file should exist")
			defer func() { _ = os.Remove(cf) }()

			got, rerr := ioutil.ReadFile(cf)
			require.NoError(t, rerr, "loading config from file shouldn't error")

			require.Equal(t, want, string(got), "config didn't match")
		})
	}
}
