package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Graph interface {
	Extend(vertex *model.Vertex) error
	Confirm(vertexID model.Hash) error
	Final() (*model.Vertex, bool)
}
