package fixture

import (
	"math/rand"
	"testing"

	"github.com/alvalor/consensus/model/base"
)

func Vertex(t testing.TB, options ...func(*base.Vertex)) *base.Vertex {
	height := rand.Uint64()
	vertex := base.Vertex{
		Height:     height,
		ParentID:   Hash(t),
		ProposerID: Hash(t),
		ArcID:      Hash(t),
	}
	for _, option := range options {
		option(&vertex)
	}
	return &vertex
}

func WithParent(parent *base.Vertex) func(*base.Vertex) {
	return func(vertex *base.Vertex) {
		vertex.Height = parent.Height + 1
		vertex.ParentID = parent.ID()
	}
}

func WithProposer(proposerID base.Hash) func(*base.Vertex) {
	return func(vertex *base.Vertex) {
		vertex.ProposerID = proposerID
	}
}
