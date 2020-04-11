package fixture

import (
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

func Proposal(t require.TestingT, options ...func(*message.Proposal)) *message.Proposal {
	proposal := message.Proposal{
		Block:     Block(t),
		Signature: Sig(t),
	}
	for _, option := range options {
		option(&proposal)
	}
	return &proposal
}

func WithBlock(block *model.Block) func(*message.Proposal) {
	return func(proposal *message.Proposal) {
		proposal.Block = block
	}
}
