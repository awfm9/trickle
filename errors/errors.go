package errors

import (
	"fmt"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type ConflictingProposal struct {
	Proposal *message.Proposal
	Final    uint64
}

func (cv ConflictingProposal) Error() string {
	return fmt.Sprintf("conflicting proposal (height: %d, final: %d)", cv.Proposal.Height, cv.Final)
}

type ObsoleteProposal struct {
	Proposal *message.Proposal
	Cutoff   uint64
}

func (op ObsoleteProposal) Error() string {
	return fmt.Sprintf("obsolete proposal (height: %d, cutoff: %d)", op.Proposal.Height, op.Cutoff)
}

type ConflictingVote struct {
	Vote  *message.Vote
	Final uint64
}

func (cv ConflictingVote) Error() string {
	return fmt.Sprintf("conflicting vote (height: %d, final: %d)", cv.Vote.Height, cv.Final)
}

type ObsoleteVote struct {
	Vote   *message.Vote
	Cutoff uint64
}

func (ov ObsoleteVote) Error() string {
	return fmt.Sprintf("obsolete vote (height: %d, cutoff: %d)", ov.Vote.Height, ov.Cutoff)
}

type InvalidProposer struct {
	Proposal *message.Proposal
	Leader   model.Hash
}

func (ip InvalidProposer) Error() string {
	return fmt.Sprintf("invalid proposer (proposer: %x, leader: %x)", ip.Proposal.SignerID, ip.Leader)
}

type InvalidCollector struct {
	Vote      *message.Vote
	Receiver  model.Hash
	Collector model.Hash
}

func (ic InvalidCollector) Error() string {
	return fmt.Sprintf("invalid collector (receiver: %x, collector: %x)", ic.Receiver, ic.Collector)
}

type InvalidSignature struct {
	Entity string
	Signer model.Hash
}

func (is InvalidSignature) Error() string {
	return fmt.Sprintf("invalid signature (entity: %s, signer: %x)", is.Entity, is.Signer)
}

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
