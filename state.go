package consensus

import (
	"github.com/alvalor/consensus/model"
)

type State interface {
	Leader(height uint64) (model.Hash, error)
	Threshold() (uint, error)
}
