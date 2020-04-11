package integration

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/test/fixture"
)

func TestSingularSet(t *testing.T) {

	// create a single participant
	p := NewParticipant(t, fixture.Hash(t), Or(AfterRound(1000)))

	// bootstrap with genesis block
	genesis := fixture.Genesis(t)
	err := p.pro.Bootstrap(genesis)
	require.NoError(t, err, "bootstrap should pass")

	// run until stop condition
	err = p.Run()
	require.NoError(t, err, "run should pass")
}

func TestMinimalSet(t *testing.T) {

	// number of nodes
	n := 3

	// create the participants to stop of 10000 blocks or error
	participants := make([]*Participant, n)
	for i := 0; i < n; i++ {
		participants[i] = NewParticipant(t, fixture.Hash(t), AfterRound(1000))
	}

	// connect all participants together
	Network(participants)

	// bootstrap each participant
	genesis := fixture.Genesis(t)
	for _, p := range participants {
		err := p.pro.Bootstrap(genesis)
		require.NoError(t, err, "bootstrap should pass")
	}

	// start execution on each participant
	var wg sync.WaitGroup
	for i := range participants {
		p := participants[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := p.Run()
			require.NoError(t, err, "run should pass")
		}()
	}
	wg.Wait()
}

func TestSmallSet(t *testing.T) {

	// number of nodes
	n := 7

	// create the participants to stop of 10000 blocks or error
	participants := make([]*Participant, n)
	for i := 0; i < n; i++ {
		participants[i] = NewParticipant(t, fixture.Hash(t), AfterRound(1000))
	}

	// connect all participants together
	Network(participants)

	// bootstrap each participant
	genesis := fixture.Genesis(t)
	for _, p := range participants {
		err := p.pro.Bootstrap(genesis)
		require.NoError(t, err, "genesis block should not error")
	}

	// start execution on each participant
	var wg sync.WaitGroup
	for i := range participants {
		p := participants[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := p.Run()
			require.NoError(t, err, "run should pass")
		}()
	}
	wg.Wait()
}

func TestBigSet(t *testing.T) {

	// number of nodes
	n := 101

	// create the participants to stop of 10000 blocks or error
	participants := make([]*Participant, n)
	for i := 0; i < n; i++ {
		participants[i] = NewParticipant(t, fixture.Hash(t), AfterRound(1000))
	}

	// connect all participants together
	Network(participants)

	// bootstrap each participant
	genesis := fixture.Genesis(t)
	for _, p := range participants {
		err := p.pro.Bootstrap(genesis)
		require.NoError(t, err, "genesis block should not error")
	}

	// start execution on each participant
	var wg sync.WaitGroup
	for i := range participants {
		p := participants[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := p.Run()
			require.NoError(t, err, "run should pass")
		}()
	}
	wg.Wait()
}
