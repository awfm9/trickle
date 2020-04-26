package fixture

import (
	"math/rand"
	"testing"

	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
)

func Vote(t testing.TB, options ...func(*message.Vote)) *message.Vote {
	vote := message.Vote{
		Height:      rand.Uint64(),
		CandidateID: Hash(t),
		SignerID:    Hash(t),
		Signature:   Sig(t),
	}
	return &vote
}

func ForCandidate(candidate *base.Vertex) func(*message.Vote) {
	return func(vote *message.Vote) {
		vote.Height = candidate.Height
		vote.CandidateID = candidate.ID()
	}
}

func WithVoter(voterID base.Hash) func(*message.Vote) {
	return func(vote *message.Vote) {
		vote.SignerID = voterID
	}
}
