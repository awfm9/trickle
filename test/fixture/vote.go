package fixture

import (
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

func Vote(t require.TestingT, options ...func(*message.Vote)) *message.Vote {
	vote := message.Vote{
		BlockID:   Hash(t),
		SignerID:  Hash(t),
		Signature: Sig(t),
	}
	return &vote
}

func WithCandidate(block *model.Block) func(*message.Vote) {
	return func(vote *message.Vote) {
		vote.BlockID = block.ID()
	}
}

func WithVoter(voterID model.Hash) func(*message.Vote) {
	return func(vote *message.Vote) {
		vote.SignerID = voterID
	}
}
