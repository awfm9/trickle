package cache

import (
	"fmt"

	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
	"github.com/alvalor/consensus/model/signal"
)

// Volatile stores proposals by height at the first level and by proposer
// at the second level. It detects two different proposal's by the same signer
// at the same height.
type Volatile struct {
	proposalLookups map[uint64](map[base.Hash]*message.Proposal)
	voteLookups     map[uint64](map[base.Hash]*message.Vote)
}

// NewVolatile creates a new proposal cache with initialized map.
func NewVolatile() *Volatile {
	v := Volatile{
		proposalLookups: make(map[uint64](map[base.Hash]*message.Proposal)),
		voteLookups:     make(map[uint64](map[base.Hash]*message.Vote)),
	}
	return &v
}

// Proposal will store the proposal for the height at which it was made. We don't
// check if multiple people propose at a height, because we alredy check that
// in our business logic and create an invalid proposer error.
func (v *Volatile) Proposal(proposal *message.Proposal) error {

	// get the proposal lookup for signers who have proposed at this height
	proposalLookup, exists := v.proposalLookups[proposal.Candidate.Height]
	if !exists {
		proposalLookup = make(map[base.Hash]*message.Proposal)
		v.proposalLookups[proposal.Candidate.Height] = proposalLookup
	}

	// if the proposer has already proposed at this height, and the vertex
	// doesn't match, we have a double proposal error for this signer
	duplicate, hasProposed := proposalLookup[proposal.Candidate.ProposerID]
	if hasProposed && duplicate.Candidate.ID() != proposal.Candidate.ID() {
		return signal.DoubleProposal{First: duplicate, Second: proposal}
	}

	// otherwise, storing is an idempotent operation
	proposalLookup[proposal.Candidate.ProposerID] = proposal
	return nil
}

// Vote will store the vote. Currently, we group them by height, as there
// should be no vertex mismatch between votes at the same height. Otherwise, the
// leader for the round send a double proposal, which should have been detected.
func (v *Volatile) Vote(vote *message.Vote) error {

	// get the vote lookup for signers who have voted at this height
	voteLookup, exists := v.voteLookups[vote.Height]
	if !exists {
		voteLookup = make(map[base.Hash]*message.Vote)
		v.voteLookups[vote.Height] = voteLookup
	}

	// if the voter has already voted on this height, and the vertex doesn't
	// match, we have a double vote error for this signer
	duplicate, hasVoted := voteLookup[vote.SignerID]
	if hasVoted && duplicate.CandidateID != vote.CandidateID {
		return signal.DoubleVote{First: duplicate, Second: vote}
	}

	// otherwise, storing is an idempotent operation
	voteLookup[vote.SignerID] = vote
	return nil
}

// Quorum will build a reference for the given height, if possible.
func (v *Volatile) Quorum(height uint64, candidateID base.Hash) (*message.Quorum, error) {

	// get the votes registered for this height
	voteLookup, exists := v.voteLookups[height]
	if !exists {
		return nil, fmt.Errorf("height unknown (%x)", height)
	}

	// get the votes registered for this vertex
	byCandidate := make(map[base.Hash][]*message.Vote)
	for _, vote := range voteLookup {
		byCandidate[vote.CandidateID] = append(byCandidate[vote.CandidateID], vote)
	}

	// create the quorum with available votes
	votes := byCandidate[candidateID]
	var signature []byte
	signerIDs := make([]base.Hash, 0, len(votes))
	for _, vote := range votes {
		signerIDs = append(signerIDs, vote.SignerID)
		signature = append(signature, vote.Signature...)
	}
	quorum := message.Quorum{
		SignerIDs: signerIDs,
		Signature: signature,
	}

	return &quorum, nil
}

// Clear will drop all cached entities at or below cutoff.
func (v *Volatile) Clear(cutoff uint64) error {
	for height := range v.proposalLookups {
		if height <= cutoff {
			delete(v.proposalLookups, height)
		}
	}
	for height := range v.voteLookups {
		if height <= cutoff {
			delete(v.voteLookups, height)
		}
	}

	return nil
}
