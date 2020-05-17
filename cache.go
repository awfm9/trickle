package consensus

import (
	"github.com/awfm/consensus/model/base"
	"github.com/awfm/consensus/model/message"
)

// Cache stores votes to build proposals.
type Cache interface {
	Proposal(proposal *message.Proposal) error
	Vote(vote *message.Vote) error
	Quorum(height uint64, vertexID base.Hash) (*message.Quorum, error)
	Clear(height uint64) error
}
