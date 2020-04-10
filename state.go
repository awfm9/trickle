package consensus

import (
	"github.com/alvalor/consensus/model"
)

type State interface {
	Height() uint64
	Set(height uint64)
	Leader(height uint64) model.Hash
	Collector(heigh uint64) model.Hash
}
