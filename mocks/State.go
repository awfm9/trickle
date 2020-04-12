// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	model "github.com/alvalor/consensus/model"
	mock "github.com/stretchr/testify/mock"
)

// State is an autogenerated mock type for the State type
type State struct {
	mock.Mock
}

// Leader provides a mock function with given fields: height
func (_m *State) Leader(height uint64) (model.Hash, error) {
	ret := _m.Called(height)

	var r0 model.Hash
	if rf, ok := ret.Get(0).(func(uint64) model.Hash); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.Hash)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Threshold provides a mock function with given fields:
func (_m *State) Threshold() (uint, error) {
	ret := _m.Called()

	var r0 uint
	if rf, ok := ret.Get(0).(func() uint); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
