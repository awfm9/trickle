package consensus

import (
	"github.com/awfm/consensus/model/base"
)

type Graph interface {
	Extend(vertex *base.Vertex) error
	Confirm(vertexID base.Hash) error
	Contains(vertexID base.Hash) (bool, error)
	Tip() (*base.Vertex, error)
	Final() (*base.Vertex, error)
}
