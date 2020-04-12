package fixture

import (
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func QC(t require.TestingT, options ...func(*model.QC)) *model.QC {
	qc := model.QC{
		Height:    rand.Uint64(),
		VertexID:  Hash(t),
		SignerIDs: Hashes(t, 3),
		Signature: Sig(t),
	}
	for _, option := range options {
		option(&qc)
	}
	return &qc
}

func WithHeight(height uint64) func(*model.QC) {
	return func(qc *model.QC) {
		qc.Height = height
	}
}
