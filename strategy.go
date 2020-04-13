package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Strategy interface {
	Threshold(height uint64) (uint, error)
	Leader(height uint64) (model.Hash, error)
	Collector(height uint64) (model.Hash, error)
}
