package model

import (
	"reflect"
	"testing"
)

var argvCases = []struct {
	in   CompilerFilter
	want []string
}{
	{CompilerFilter{}, nil},
	{CompilerFilter{CompPred: "(id (has_tag yeet))"}, []string{"-compiler-predicate", "(id (has_tag yeet))"}},
	{CompilerFilter{MachPred: "(id (has_tag yote))"}, []string{"-machine-predicate", "(id (has_tag yote))"}},
	{CompilerFilter{CompPred: "(id (has_tag yeet))", MachPred: "(id (has_tag yote))"},
		[]string{"-compiler-predicate", "(id (has_tag yeet))", "-machine-predicate", "(id (has_tag yote))"}},
}

func TestCompilerFilter_ToArgv(t *testing.T) {
	for _, c := range argvCases {
		got := c.in.ToArgv()
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("(%v).ToArgv=%v; want %v", c.in, got, c.want)
		}
	}
}
