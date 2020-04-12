package fixture

import (
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func Genesis(t require.TestingT) *model.Vertex {
	genesis := model.Vertex{
		QC:       nil,
		Height:   0,
		ArcID:    model.ZeroHash,
		SignerID: model.ZeroHash,
	}
	return &genesis
}
