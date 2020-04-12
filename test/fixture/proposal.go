package fixture

import (
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

func Proposal(t require.TestingT, options ...func(*message.Proposal)) *message.Proposal {
	proposal := message.Proposal{
		Vertex:    Vertex(t),
		Signature: Sig(t),
	}
	for _, option := range options {
		option(&proposal)
	}
	return &proposal
}

func WithVertex(vertex *model.Vertex) func(*message.Proposal) {
	return func(proposal *message.Proposal) {
		proposal.Vertex = vertex
	}
}
