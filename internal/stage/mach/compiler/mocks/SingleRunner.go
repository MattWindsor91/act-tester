// Code generated by mockery v2.1.0. DO NOT EDIT.

package mocks

import (
	compile "github.com/MattWindsor91/act-tester/internal/model/job/compile"

	context "context"

	io "io"

	mock "github.com/stretchr/testify/mock"
)

// SingleRunner is an autogenerated mock type for the SingleRunner type
type SingleRunner struct {
	mock.Mock
}

// RunCompiler provides a mock function with given fields: ctx, j, errw
func (_m *SingleRunner) RunCompiler(ctx context.Context, j compile.Single, errw io.Writer) error {
	ret := _m.Called(ctx, j, errw)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, compile.Single, io.Writer) error); ok {
		r0 = rf(ctx, j, errw)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
