package cache

import (
	"github.com/alvalor/consensus/errors"
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

// Proposals stores proposals by height at the first level and by proposer
// at the second level. It detects two different proposal's by the same signer
// at the same height.
type Proposals struct {
	proposalLookups map[uint64](map[model.Hash]*message.Proposal)
}

// ForProposals creates a new proposal cache with initialized map.
func ForProposals() *Proposals {
	pc := Proposals{
		proposalLookups: make(map[uint64](map[model.Hash]*message.Proposal)),
	}
	return &pc
}

// Store will store the proposal for the height at which it was made. We don't
// check if multiple people propose at a height, because we alredy check that
// in our business logic and create an invalid proposer error.
func (pc *Proposals) Store(proposal *message.Proposal) (bool, error) {

	// get the proposal lookup for signers who have proposed at this height
	proposalLookup, exists := pc.proposalLookups[proposal.Height]
	if !exists {
		proposalLookup = make(map[model.Hash]*message.Proposal)
		pc.proposalLookups[proposal.Height] = proposalLookup
	}

	// if the proposer has already proposed at this height, and the vertex
	// doesn't match, we have a double proposal error for this signer
	duplicate, hasProposed := proposalLookup[proposal.SignerID]
	if hasProposed && duplicate.Vertex.ID() != proposal.Vertex.ID() {
		return false, errors.DoubleProposal{First: duplicate, Second: proposal}
	}

	// otherwise, if the proposer hasn't proposed yet, we should store the proposal
	if !hasProposed {
		proposalLookup[proposal.SignerID] = proposal
		return true, nil
	}

	// the remaining path is a no-op
	return false, nil
}

// Clear will drop all proposals at or below the given cutoff.
func (pc *Proposals) Clear(cutoff uint64) error {
	for height := range pc.proposalLookups {
		if height <= cutoff {
			delete(pc.proposalLookups, height)
		}
	}
	return nil
}
