package integration

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/test/fixture"
)

func TestSingularSet(t *testing.T) {

	// create a single participant
	log := zerolog.New(os.Stderr)
	p := NewParticipant(t,
		WithLog(log),
		WithRound(0),
		WithIgnore(),
		WithStop(
			AfterRound(4096, errFinished),
			AfterDelay(8*time.Second, errTimeout),
		),
	)

	// bootstrap with genesis block
	genesis := fixture.Genesis(t)
	err := p.pro.Bootstrap(genesis)
	require.NoError(t, err, "bootstrap should pass")

	// run until stop condition
	err = p.Run()
	require.True(t, errors.Is(err, errFinished), "run should finish successfully (%s)", err)
}

func TestMinimalSet(t *testing.T) {

	// number of nodes
	n := uint(3)

	// create the participants to stop of 10000 blocks or error
	participantIDs := fixture.Hashes(t, n)
	participants := make([]*Participant, 0, len(participantIDs))
	for index, selfID := range participantIDs {
		log := zerolog.New(os.Stderr).With().
			Int("index", index).
			Hex("self", selfID[:]).
			Logger()
		p := NewParticipant(t,
			WithLog(log),
			WithSelf(selfID),
			WithParticipants(participantIDs),
			WithStop(
				AfterRound(1024, errFinished),
				AfterDelay(8*time.Second, errTimeout),
			),
		)
		participants = append(participants, p)
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
			require.True(t, errors.Is(err, errFinished), "run should finish successfully (%s)", err)
		}()
	}
	wg.Wait()
}

func TestSmallSet(t *testing.T) {

	// number of nodes
	n := uint(7)

	// create the participants to stop of 10000 blocks or error
	participantIDs := fixture.Hashes(t, n)
	participants := make([]*Participant, 0, len(participantIDs))
	for index, selfID := range participantIDs {
		log := zerolog.New(os.Stderr).With().
			Int("index", index).
			Hex("self", selfID[:]).
			Logger()
		p := NewParticipant(t,
			WithLog(log),
			WithSelf(selfID),
			WithParticipants(participantIDs),
			WithStop(
				AfterRound(512, errFinished),
				AfterDelay(8*time.Second, errTimeout),
			),
		)
		participants = append(participants, p)
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
			require.True(t, errors.Is(err, errFinished), "run should finish successfully (%s)", err)
		}()
	}
	wg.Wait()
}

func TestBigSet(t *testing.T) {

	// number of nodes
	n := uint(101)

	// create the participants to stop of 10000 blocks or error
	participantIDs := fixture.Hashes(t, n)
	participants := make([]*Participant, 0, len(participantIDs))
	for index, selfID := range participantIDs {
		log := zerolog.New(os.Stderr).With().
			Int("index", index).
			Hex("self", selfID[:]).
			Logger()
		p := NewParticipant(t,
			WithLog(log),
			WithSelf(selfID),
			WithParticipants(participantIDs),
			WithStop(
				AfterRound(32, errFinished),
				AfterDelay(8*time.Second, errTimeout),
			),
		)
		participants = append(participants, p)
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
			require.True(t, errors.Is(err, errFinished), "run should finish successfully (%s)", err)
		}()
	}
	wg.Wait()
}
