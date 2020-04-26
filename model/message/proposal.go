package message

import (
	"github.com/alvalor/consensus/model/base"
)

// Proposal is a proposal for a new vertex in the consensus graph. It contains
// the proposed vertex, a quorum for the parent graph and the signature of the
// proposer.
type Proposal struct {
	Candidate *base.Vertex
	Quorum    *Quorum
	Signature base.Signature
}

// Vote returns the vote of the proposer that is implicitly included in each
// proposal.
func (p *Proposal) Vote() *Vote {

	vote := Vote{
		Height:      p.Candidate.Height,
		CandidateID: p.Candidate.ID(),
		SignerID:    p.Candidate.ProposerID,
		Signature:   p.Signature,
	}

	return &vote
}
