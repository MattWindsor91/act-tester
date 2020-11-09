// Copyright (c) 2020 Matt Windsor and contributors
//
// This file is part of act-tester.
// Licenced under the MIT licence; see `LICENSE`.

package recipe

import (
	"strconv"
	"strings"

	"github.com/MattWindsor91/act-tester/internal/model/filekind"
)

// PopAll is the value to pass to NPops to ask the instruction to pop all applicable files off the stack.
const PopAll = 0

// Instruction represents a single instruction in a recipe.
//
// Instructions target a stack machine in the machine node.
type Instruction struct {
	// Op is the opcode.
	Op Op `json:"op"`

	// File is, if applicable, the file argument to the instruction.
	File string `json:"file,omitempty"`

	// FileKind is, if applicable, the file kind argument to the instruction.
	FileKind filekind.Kind `json:"file_kind,omitempty"`

	// NPops is, if applicable and nonzero, the maximum number of items to pop off the file stack.
	NPops int `json:"npops,omitempty"`
}

// IsRuntime gets whether this instruction is a run-time one.
// The machine node will, at time of writing, segregate compile-time and run-time instructions.
func (i Instruction) IsRuntime() bool {
	return LastCompile < i.Op
}

// String produces a human-readable string representation of this instruction.
func (i Instruction) String() string {
	strs := []string{i.Op.String()}

	switch i.Op {
	case CompileExe:
		fallthrough
	case CompileObj:
		fallthrough
	case RunExe:
		fallthrough
	case Cat:
		strs = append(strs, npopString(i.NPops))
	case PushInput:
		strs = append(strs, strconv.Quote(i.File))
	case PushInputs:
		strs = append(strs, i.FileKind.String())
	}

	return strings.Join(strs, " ")
}

// npopString returns 'ALL' if npops requests popping all files, or npops as a string otherwise.
func npopString(npops int) string {
	if npops <= PopAll {
		return "ALL"
	}
	return strconv.Itoa(npops)
}

// CompileExeInst produces a 'compile binary' instruction.
func CompileExeInst(npops int) Instruction {
	return Instruction{Op: CompileExe, NPops: npops}
}

// CompileObjInst produces a 'compile object' instruction.
func CompileObjInst(npops int) Instruction {
	return Instruction{Op: CompileObj, NPops: npops}
}

/*
// RunExe produces a 'run' instruction.
func RunExeInst(npops int) Instruction {
	return Instruction{Op: RunExe, NPops: npops}
}
*/

// CatInst produces a 'cat' instruction.
func CatInst(npops int) Instruction {
	return Instruction{Op: Cat, NPops: npops}
}

// PushInputInst produces a 'push input' instruction.
func PushInputInst(file string) Instruction {
	return Instruction{Op: PushInput, File: file}
}

// PushInputsInst produces a 'push inputs' instruction.
func PushInputsInst(kind filekind.Kind) Instruction {
	return Instruction{Op: PushInputs, FileKind: kind}
}
