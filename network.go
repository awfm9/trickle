package consensus

import (
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Network interface {
	Broadcast(proposal *message.Proposal) error
	Transmit(vote *message.Vote, recipientID model.Hash) error
}
