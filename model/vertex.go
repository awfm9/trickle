package model

import (
	"encoding/json"

	"golang.org/x/crypto/sha3"
)

// Vertex represens one node in the directed graph of the consensus algorithm.
type Vertex struct {
	Parent   *Reference // parent is the reference to the parent & its confirmation
	Height   uint64     // heigth is how far the vertex is removed from the source
	ArcID    Hash       // arc represents the edge between parent and child vertex
	SignerID Hash       // signer is the proposer of this vertex (should be leader)
}

// ID returns a unique identifier for the given vertex.
func (v Vertex) ID() Hash {
	data, _ := json.Marshal(v)
	hash := sha3.Sum256(data)
	return hash
}
