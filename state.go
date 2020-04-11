package consensus

import (
	"github.com/alvalor/consensus/model"
)

type State interface {
	Round() (uint64, error)
	Set(height uint64) error
	Leader(height uint64) (model.Hash, error)
	Threshold() (uint, error)
}
