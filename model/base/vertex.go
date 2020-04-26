package base

import (
	"encoding/json"

	"golang.org/x/crypto/sha3"
)

// Vertex is a node in the directed graph of the consensus algorithm.
type Vertex struct {
	Height     uint64 // height is how far the vertex is removed from the source
	ParentID   Hash   // parent is the reference to the parent & its confirmation
	ProposerID Hash   // the proposer of this vertex (should be leader at height)
	ArcID      Hash   // arc represents the edge between parent and child vertex
}

// ID returns a unique identifier for the given vertex.
func (v Vertex) ID() Hash {
	data, _ := json.Marshal(v)
	hash := sha3.Sum256(data)
	return hash
}
