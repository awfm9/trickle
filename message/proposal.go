package message

import (
	"github.com/alvalor/consensus/model"
)

// Proposal is a proposal for a given consensus rounds. It contains the proposed
// vertex candidate, along with the quorum certificate that confirms the majority
// consensus for its parent.
type Proposal struct {
	*model.Vertex
	Signature model.Signature
}

// Vote extracts the vote of the proposer that is included implicitly with every
// proposal.
func (p *Proposal) Vote() *Vote {
	vote := Vote{
		Height:    p.Height,
		VertexID:  p.Vertex.ID(),
		SignerID:  p.SignerID,
		Signature: p.Signature,
	}
	return &vote
}
