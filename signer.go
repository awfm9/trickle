package consensus

import (
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Signer interface {
	Vote(block *model.Block) (*message.Vote, error)
	Propose(block *model.Block) (*message.Proposal, error)
}
