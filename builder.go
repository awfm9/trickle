package consensus

import (
	"github.com/awfm/consensus/model/base"
)

type Builder interface {
	Arc() (base.Hash, error)
}
