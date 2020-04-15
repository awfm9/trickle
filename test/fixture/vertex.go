package fixture

import (
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func Vertex(t require.TestingT, options ...func(*model.Vertex)) *model.Vertex {
	height := rand.Uint64()
	vertex := model.Vertex{
		Height:   height,
		Parent:   Reference(t, WithHeight(height-1)),
		ArcID:    Hash(t),
		SignerID: Hash(t),
	}
	for _, option := range options {
		option(&vertex)
	}
	return &vertex
}

func WithParent(parent *model.Vertex) func(*model.Vertex) {
	return func(vertex *model.Vertex) {
		vertex.Parent.Height = parent.Height
		vertex.Parent.VertexID = parent.ID()
		vertex.Height = parent.Height + 1
	}
}

func WithProposer(proposerID model.Hash) func(*model.Vertex) {
	return func(vertex *model.Vertex) {
		vertex.SignerID = proposerID
	}
}
