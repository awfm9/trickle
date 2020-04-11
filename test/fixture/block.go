package fixture

import (
	"math/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func Block(t require.TestingT, options ...func(*model.Block)) *model.Block {
	block := model.Block{
		Height:      rand.Uint64(),
		QC:          QC(t),
		PayloadHash: Hash(t),
		Timestamp:   time.Now().UTC(),
		SignerID:    Hash(t),
	}
	for _, option := range options {
		option(&block)
	}
	return &block
}

func WithParent(parent *model.Block) func(*model.Block) {
	return func(block *model.Block) {
		block.QC.BlockID = parent.ID()
		block.Height = parent.Height - 1
	}
}

func WithProposer(proposerID model.Hash) func(*model.Block) {
	return func(block *model.Block) {
		block.SignerID = proposerID
	}
}
