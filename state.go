package consensus

import (
	"github.com/alvalor/consensus/model"
)

type State interface {
	Round() uint64
	Set(height uint64)
	Leader(height uint64) model.Hash
	Threshold() uint
}
