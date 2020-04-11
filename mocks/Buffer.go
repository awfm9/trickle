// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	message "github.com/alvalor/consensus/message"
	mock "github.com/stretchr/testify/mock"

	model "github.com/alvalor/consensus/model"
)

// Buffer is an autogenerated mock type for the Buffer type
type Buffer struct {
	mock.Mock
}

// Clear provides a mock function with given fields: blockID
func (_m *Buffer) Clear(blockID model.Hash) {
	_m.Called(blockID)
}

// Tally provides a mock function with given fields: vote
func (_m *Buffer) Tally(vote *message.Vote) error {
	ret := _m.Called(vote)

	var r0 error
	if rf, ok := ret.Get(0).(func(*message.Vote) error); ok {
		r0 = rf(vote)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Votes provides a mock function with given fields: blockID
func (_m *Buffer) Votes(blockID model.Hash) []*message.Vote {
	ret := _m.Called(blockID)

	var r0 []*message.Vote
	if rf, ok := ret.Get(0).(func(model.Hash) []*message.Vote); ok {
		r0 = rf(blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*message.Vote)
		}
	}

	return r0
}
