// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package model_test

import (
	"fmt"

	"github.com/MattWindsor91/act-tester/internal/pkg/model"
)

// ExampleObs_Sat is a testable example for Sat.
func ExampleObs_Sat() {
	o1 := model.Obs{}
	fmt.Println("o1:", o1.Sat())

	o2 := model.Obs{Flags: model.ObsUnsat}
	fmt.Println("o2:", o2.Sat())

	o3 := model.Obs{Flags: model.ObsSat}
	fmt.Println("o3:", o3.Sat())

	// Output:
	// o1: false
	// o2: false
	// o3: true
}

// ExampleObs_Unsat is a testable example for Unsat.
func ExampleObs_Unsat() {
	o1 := model.Obs{}
	fmt.Println("o1:", o1.Unsat())

	o2 := model.Obs{Flags: model.ObsUnsat}
	fmt.Println("o2:", o2.Unsat())

	o3 := model.Obs{Flags: model.ObsSat}
	fmt.Println("o3:", o3.Unsat())

	// Output:
	// o1: false
	// o2: true
	// o3: false
}
