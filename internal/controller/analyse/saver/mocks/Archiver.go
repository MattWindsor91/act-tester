// Code generated by mockery v1.1.2. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Archiver is an autogenerated mock type for the Archiver type
type Archiver struct {
	mock.Mock
}

// ArchiveFile provides a mock function with given fields: rpath, wpath, mode
func (_m *Archiver) ArchiveFile(rpath string, wpath string, mode int64) error {
	ret := _m.Called(rpath, wpath, mode)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, int64) error); ok {
		r0 = rf(rpath, wpath, mode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *Archiver) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}