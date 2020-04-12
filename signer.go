package consensus

import (
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Signer interface {
	Self() (model.Hash, error)
	Proposal(vertex *model.Vertex) (*message.Proposal, error)
	Vote(vertex *model.Vertex) (*message.Vote, error)
}
