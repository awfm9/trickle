package consensus

import (
	"github.com/alvalor/consensus/event"
	"github.com/alvalor/consensus/model"
)

type Signer interface {
	Vote(block *model.Block) *event.Vote
	Propose(block *model.Block) *event.Proposal
}
