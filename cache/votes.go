package cache

import (
	"fmt"

	"github.com/alvalor/consensus/errors"
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

// Votes keeps track of votes per height on the first level, and votes per
// signer on the second level. It detects double votes of the same signer on
// two different vertices at the same height.
type Votes struct {
	voteLookups map[uint64](map[model.Hash]*message.Vote)
}

// ForVotes creates a new vote cache with initialized map.
func ForVotes() *Votes {
	vc := Votes{
		voteLookups: make(map[uint64](map[model.Hash]*message.Vote)),
	}
	return &vc
}

// Store will store the vote. Currently, we group them by height, as there
// should be no vertex mismatch between votes at the same height. Otherwise, the
// leader for the round send a double proposal, which should have been detected.
func (vc *Votes) Store(vote *message.Vote) (bool, error) {

	// get the vote lookup for signers who have voted at this height
	voteLookup, exists := vc.voteLookups[vote.Height]
	if !exists {
		voteLookup = make(map[model.Hash]*message.Vote)
		vc.voteLookups[vote.Height] = voteLookup
	}

	// if the voter has already voted on this height, and the vertex doesn't
	// match, we have a double vote error for this signer
	duplicate, hasVoted := voteLookup[vote.SignerID]
	if hasVoted && duplicate.VertexID != vote.VertexID {
		return false, errors.DoubleVote{First: duplicate, Second: vote}
	}

	// otherwise, if the voter hasn't voted yet, we should store the vote
	if !hasVoted {
		voteLookup[vote.SignerID] = vote
		return true, nil
	}

	// the remaining path is a no-op
	return false, nil
}

// Retrieve gets the votes at a given height for a given vertex.
func (vc *Votes) Retrieve(height uint64) (model.Hash, []*message.Vote, error) {

	// get the votes registered for this height
	voteLookup, exists := vc.voteLookups[height]
	if !exists {
		return model.ZeroHash, nil, fmt.Errorf("height unknown (%x)", height)
	}

	// add the votes that have the desired vertex ID to a slice
	byVertex := make(map[model.Hash][]*message.Vote)
	for _, vote := range voteLookup {
		byVertex[vote.VertexID] = append(byVertex[vote.VertexID], vote)
	}

	// find the vertex with most votes
	max := 0
	bestID := model.ZeroHash
	for vertexID, votes := range byVertex {
		if len(votes) > max {
			bestID = vertexID
			max = len(votes)
		}
	}

	return bestID, byVertex[bestID], nil
}

// Clear will drop all votes at or below the given cutoff, regardless of vertex.
func (vc *Votes) Clear(cutoff uint64) error {
	for height := range vc.voteLookups {
		if height <= cutoff {
			delete(vc.voteLookups, height)
		}
	}
	return nil
}
