package errors

import (
	"fmt"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

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
