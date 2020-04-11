package integration

import (
	"testing"

	"github.com/alvalor/consensus/test/fixture"
	"github.com/stretchr/testify/require"
)

func TestFullSet(t *testing.T) {

	// number of nodes
	n := 3

	// create the participants to stop of 10000 blocks or error
	participants := make([]*Participant, n)
	for i := 0; i < n; i++ {
		participants[i] = NewParticipant(t, fixture.Hash(t), AtRound(1000))
	}

	// connect all participants together
	Network(participants)

	// bootstrap each participant
	genesis := fixture.Genesis(t)
	for _, p := range participants {
		err := p.pro.Bootstrap(genesis)
		require.NoError(t, err, "genesis block should not error")
	}

	// wait for each participant to end and check no error occured
	for _, p := range participants {
		p.wait()
		require.NoError(t, p.last, "no error should have occurred")
	}
}
