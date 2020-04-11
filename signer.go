package consensus

import (
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Signer interface {
	Proposal(block *model.Block) (*message.Proposal, error)
	Vote(block *model.Block) (*message.Vote, error)
}
