package fixture

import (
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

func Vote(t require.TestingT, options ...func(*message.Vote)) *message.Vote {
	vote := message.Vote{
		Height:    rand.Uint64(),
		VertexID:  Hash(t),
		SignerID:  Hash(t),
		Signature: Sig(t),
	}
	return &vote
}

func WithCandidate(vertex *model.Vertex) func(*message.Vote) {
	return func(vote *message.Vote) {
		vote.Height = vertex.Height
		vote.VertexID = vertex.ID()
	}
}

func WithVoter(voterID model.Hash) func(*message.Vote) {
	return func(vote *message.Vote) {
		vote.SignerID = voterID
	}
}
