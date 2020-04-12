package fixture

import (
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func Parent(t require.TestingT, options ...func(*model.Parent)) *model.Parent {
	parent := model.Parent{
		Height:    rand.Uint64(),
		VertexID:  Hash(t),
		SignerIDs: Hashes(t, 3),
		Signature: Sig(t),
	}
	for _, option := range options {
		option(&parent)
	}
	return &parent
}

func WithHeight(height uint64) func(*model.Parent) {
	return func(parent *model.Parent) {
		parent.Height = height
	}
}
