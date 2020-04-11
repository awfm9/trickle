// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	message "github.com/alvalor/consensus/message"
	mock "github.com/stretchr/testify/mock"
)

// Verifier is an autogenerated mock type for the Verifier type
type Verifier struct {
	mock.Mock
}

// Proposal provides a mock function with given fields: proposal
func (_m *Verifier) Proposal(proposal *message.Proposal) error {
	ret := _m.Called(proposal)

	var r0 error
	if rf, ok := ret.Get(0).(func(*message.Proposal) error); ok {
		r0 = rf(proposal)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Vote provides a mock function with given fields: vote
func (_m *Verifier) Vote(vote *message.Vote) error {
	ret := _m.Called(vote)

	var r0 error
	if rf, ok := ret.Get(0).(func(*message.Vote) error); ok {
		r0 = rf(vote)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
