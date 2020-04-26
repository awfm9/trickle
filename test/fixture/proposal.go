package fixture

import (
	"testing"

	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
)

func Proposal(t testing.TB, options ...func(*message.Proposal)) *message.Proposal {
	proposal := message.Proposal{
		Candidate: Vertex(t),
		Signature: Sig(t),
	}
	for _, option := range options {
		option(&proposal)
	}
	return &proposal
}

func WithCandidate(candidate *base.Vertex) func(*message.Proposal) {
	return func(proposal *message.Proposal) {
		proposal.Candidate = candidate
	}
}
