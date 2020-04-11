package integration

import (
	"testing"

	"github.com/alvalor/consensus/test/fixture"
	"github.com/stretchr/testify/require"
)

func TestSingleNode(t *testing.T) {
	p := NewParticipant(t, fixture.Hash(t))
	genesis := fixture.Genesis(t)
	err := p.pro.Bootstrap(genesis)
	require.NoError(t, err, "genesis block should not error")
}
