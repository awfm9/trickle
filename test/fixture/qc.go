package fixture

import (
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model"
)

func QC(t require.TestingT) *model.QC {
	qc := model.QC{
		BlockID:   Hash(t),
		SignerIDs: Hashes(t, 3),
		Signature: Sig(t),
	}
	return &qc
}
