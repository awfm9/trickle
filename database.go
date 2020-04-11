package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Database interface {
	Block(blockID model.Hash) (*model.Block, error)
}
