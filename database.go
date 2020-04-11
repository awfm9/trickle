package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Database interface {
	Store(block *model.Block) error
	Block(blockID model.Hash) (*model.Block, error)
}
