package signal

import (
	"fmt"

	"github.com/awfm/consensus/model/base"
	"github.com/awfm/consensus/model/message"
)

// StaleProposal is an error returned when skipping processing of a proposal
// because it has already been included in our graph state.
type StaleProposal struct {
	Proposal *message.Proposal
}

func (sp StaleProposal) Error() string {
	return fmt.Sprintf("stale proposal (height: %d, candidate: %x)", sp.Proposal.Candidate.Height, sp.Proposal.Candidate.ID())
}

// InvalidProposer is an error that returned when processing a proposal made by
// a proposer who is not the leader for the proposal round.
type InvalidProposer struct {
	Proposal *message.Proposal
	Leader   base.Hash
}

func (ip InvalidProposer) Error() string {
	return fmt.Sprintf("invalid proposer (proposer: %x, leader: %x)", ip.Proposal.Candidate.ProposerID, ip.Leader)
}

// ConflictingProposal is an error returned when processing a proposal that is
// in conflict with the immutable finalized graph state.
type ConflictingProposal struct {
	Proposal *message.Proposal
	Final    *base.Vertex
}

func (cp ConflictingProposal) Error() string {
	return fmt.Sprintf("conflicting proposal (height: %d, final: %d)", cp.Proposal.Candidate.Height, cp.Final.Height)
}

// ObsoleteProposal is an error returned when processing a proposal that is
// already behind a pending proposal that the majority of the network has agreed
// on.
type ObsoleteProposal struct {
	Proposal *message.Proposal
	Tip      *base.Vertex
}

func (op ObsoleteProposal) Error() string {
	return fmt.Sprintf("obsolete proposal (height: %d, tip: %d)", op.Proposal.Candidate.Height, op.Tip.Height)
}

// DoubleProposal is an error that is returned when trying to store a proposal
// by a proposer who has already made a different proposal for the same height.
type DoubleProposal struct {
	First  *message.Proposal
	Second *message.Proposal
}

func (dp DoubleProposal) Error() string {
	return fmt.Sprintf("double proposal (height: %d, proposer: %x, proposal1: %x, proposal2: %x)", dp.First.Candidate.Height, dp.First.Candidate.ProposerID, dp.First.Candidate.ID(), dp.Second.Candidate.ID())
}
