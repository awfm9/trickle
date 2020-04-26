package consensus

import (
	"github.com/alvalor/consensus/model/base"
)

type Strategy interface {
	Threshold(height uint64) (uint, error)
	Leader(height uint64) (base.Hash, error)
	Collector(height uint64) (base.Hash, error)
}
