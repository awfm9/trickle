package errors

import (
	"fmt"

	"github.com/alvalor/consensus/message"
)

type ConflictingProposal struct {
	Proposal *message.Proposal
	Final    uint64
}

func (cv ConflictingProposal) Error() string {
	return fmt.Sprintf("conflicting proposal (height: %d, final: %d)", cv.Proposal.Height, cv.Final)
}

type ConflictingVote struct {
	Vote  *message.Vote
	Final uint64
}

func (cv ConflictingVote) Error() string {
	return fmt.Sprintf("conflicting vote (height: %d, final: %d)", cv.Vote.Height, cv.Final)
}
