package fixture

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func Genesis(t require.TestingT) *model.Block {
	genesis := model.Block{
		Height:      0,
		QC:          nil,
		PayloadHash: model.ZeroHash,
		Timestamp:   time.Now().UTC(),
		SignerID:    model.ZeroHash,
	}
	return &genesis
}
