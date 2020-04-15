package fixture

import (
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func Reference(t require.TestingT, options ...func(*model.Reference)) *model.Reference {
	parent := model.Reference{
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

func WithHeight(height uint64) func(*model.Reference) {
	return func(parent *model.Reference) {
		parent.Height = height
	}
}
