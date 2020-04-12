package integration

import (
	"github.com/stretchr/testify/require"
)

func Bootstrap(t require.TestingT, participants ...*Participant) {
	for _, p := range participants {
		err := p.pro.Bootstrap(p.genesisID)
		require.NoError(t, err, "should be able to bootstrap")
	}
}
