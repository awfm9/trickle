package errors

import (
	"fmt"

	"github.com/alvalor/consensus/message"
)

type ObsoleteProposal struct {
	Proposal *message.Proposal
	Cutoff   uint64
}

func (op ObsoleteProposal) Error() string {
	return fmt.Sprintf("obsolete proposal (height: %d, cutoff: %d)", op.Proposal.Height, op.Cutoff)
}

type ObsoleteVote struct {
	Vote   *message.Vote
	Cutoff uint64
}

func (ov ObsoleteVote) Error() string {
	return fmt.Sprintf("obsolete vote (height: %d, cutoff: %d)", ov.Vote.Height, ov.Cutoff)
}
