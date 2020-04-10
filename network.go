package consensus

import (
	"github.com/alvalor/consensus/event"
	"github.com/alvalor/consensus/model"
)

type Network interface {
	Broadcast(proposal *event.Proposal) error
	Transmit(vote *event.Vote, recipient model.Hash) error
}
