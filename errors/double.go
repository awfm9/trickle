package errors

import (
	"fmt"

	"github.com/alvalor/consensus/message"
)

type DoubleProposal struct {
	First  *message.Proposal
	Second *message.Proposal
}

func (dp DoubleProposal) Error() string {
	return fmt.Sprintf("double proposal (proposer: %x, vertex1: %x, vertex2: %x)", dp.First.SignerID, dp.First.ID(), dp.Second.ID())
}

type DoubleVote struct {
	First  *message.Vote
	Second *message.Vote
}

func (dv DoubleVote) Error() string {
	return fmt.Sprintf("double vote (voter: %x, vertex1: %x, vertex2: %x)", dv.First.SignerID, dv.First.VertexID, dv.Second.VertexID)
}
