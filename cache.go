package consensus

import (
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

// VoteCache stores votes to build proposals.
type VoteCache interface {
	Store(vote *message.Vote) (bool, error)
	Retrieve(height uint64, vertexID model.Hash) ([]*message.Vote, error)
	Clear(height uint64) error
}

// ProposalCache stores competing proposal at a given height.
type ProposalCache interface {
	Store(proposal *message.Proposal) (bool, error)
	Clear(height uint64) error
}
