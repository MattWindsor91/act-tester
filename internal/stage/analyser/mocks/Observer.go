// Code generated by mockery v2.1.0. DO NOT EDIT.

package mocks

import (
	analyser "github.com/MattWindsor91/act-tester/internal/plan/analyser"
	mock "github.com/stretchr/testify/mock"
)

// Observer is an autogenerated mock type for the Observer type
type Observer struct {
	mock.Mock
}

// OnAnalysis provides a mock function with given fields: a
func (_m *Observer) OnAnalysis(a analyser.Analysis) {
	_m.Called(a)
}
