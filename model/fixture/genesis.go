package fixture

import (
	"testing"

	"github.com/awfm/consensus/model/base"
)

func Genesis(t testing.TB) *base.Vertex {
	genesis := base.Vertex{
		Height:     0,
		ParentID:   base.ZeroHash,
		ArcID:      base.ZeroHash,
		ProposerID: base.ZeroHash,
	}
	return &genesis
}
