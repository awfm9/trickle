package consensus

import (
	"github.com/alvalor/consensus/model"
)

type Graph interface {
	Extend(vertex *model.Vertex) error
	Confirm(vertexID model.Hash) error
	Contains(vertexID model.Hash) (bool, error)
	Tip() (*model.Vertex, error)
	Final() (*model.Vertex, error)
}
