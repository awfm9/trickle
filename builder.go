package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Builder interface {
	Arc() (model.Hash, error)
}
