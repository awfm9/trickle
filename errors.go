package consensus

import (
	"fmt"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
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

type InvalidSigner struct {
	Entity string
	Signer model.Hash
}

func (is InvalidSigner) Error() string {
	return fmt.Sprintf("invalid signer (entity: %s, signer: %x)", is.Signer)
}

type InvalidSignature struct {
	Signer model.Hash
}

func (is InvalidSignature) Error() string {
	return fmt.Sprintf("invalid signature (entity: %s, signer: %x)", is.Signer)
}

type DoubleProposal struct {
	First  *message.Proposal
	Second *message.Proposal
}

func (dp DoubleProposal) Error() string {
	return fmt.Sprintf("double proposal (proposer: %x, block1: %x, block2: %x)", dp.First.SignerID, dp.First.ID(), dp.Second.ID())
}

type DoubleVote struct {
	First  *message.Vote
	Second *message.Vote
}

func (dv DoubleVote) Error() string {
	return fmt.Sprintf("double vote (voter: %x, block1: %x, block2: %x)", dv.First.SignerID, dv.First.BlockID, dv.Second.BlockID)
}
