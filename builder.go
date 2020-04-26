package consensus

import (
	"github.com/alvalor/consensus/model/base"
)

type Builder interface {
	Arc() (base.Hash, error)
}
