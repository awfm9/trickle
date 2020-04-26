package consensus

import (
	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
)

type Network interface {
	Broadcast(proposal *message.Proposal) error
	Transmit(vote *message.Vote, recipientID base.Hash) error
}
