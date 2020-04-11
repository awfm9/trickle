package integration

import (
	"testing"

	"github.com/alvalor/consensus/test/fixture"
	"github.com/stretchr/testify/require"
)

func TestSingleNode(t *testing.T) {

	// create the participant to stop at 10000 blocks
	p := NewParticipant(t, fixture.Hash(t),
		Or(AtRound(10000), Error()),
	)

	genesis := fixture.Genesis(t)
	err := p.pro.Bootstrap(genesis)
	require.NoError(t, err, "genesis block should not error")

	p.wait()
	require.NoError(t, p.last, "no error should have occurred")
}
