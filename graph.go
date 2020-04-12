package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Graph interface {
	Round() (uint64, error)
	Extend(vertex *model.Vertex) error
	Confirm(vertexID model.Hash) error
}
