package consensus

import (
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Buffer interface {
	Proposal(proposal *message.Proposal) (bool, error)
	Vote(vote *message.Vote) (bool, error)
	Votes(blockID model.Hash) ([]*message.Vote, error)
	Clear(height uint64) error
}
