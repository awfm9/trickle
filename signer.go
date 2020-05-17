package consensus

import (
	"github.com/awfm/consensus/model/base"
	"github.com/awfm/consensus/model/message"
)

type Signer interface {
	Self() (base.Hash, error)
	Proposal(vertex *base.Vertex) (*message.Proposal, error)
	Vote(vertex *base.Vertex) (*message.Vote, error)
}
