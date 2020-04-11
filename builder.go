package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Builder interface {
	PayloadHash() (model.Hash, error)
}
