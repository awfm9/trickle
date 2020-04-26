package integration

import (
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/test/fixture"
)

// TODO: fix these after processor changes

func TestSingularSet(t *testing.T) {

	// create a single participant
	log := zerolog.New(os.Stderr)
	p := NewParticipant(t,
		WithLog(log),
		WithGenesis(base.ZeroHash),
		WithIgnore(),
		WithStop(
			AfterFinal(2048, errFinished),
			AfterDelay(8*time.Second, errTimeout),
		),
	)

	// bootstrap the participant
	Bootstrap(t, p)

	// run until stop condition
	err := p.Run()
	require.True(t, errors.Is(err, errFinished), "run should finish successfully (%s)", err)
}

func TestMinimalSet(t *testing.T) {

	// number of nodes
	n := uint(3)

	// create the participants
	participantIDs := fixture.Hashes(t, n)
	participants := make([]*Participant, 0, len(participantIDs))
	for index, selfID := range participantIDs {
		log := zerolog.New(os.Stderr).With().
			Timestamp().
			Int("index", index).
			Hex("self", selfID[:]).
			Logger()
		p := NewParticipant(t,
			WithLog(log),
			WithSelf(selfID),
			WithParticipants(participantIDs),
			WithStop(
				AfterFinal(1024, errFinished),
				AfterDelay(8*time.Second, errTimeout),
			),
		)
		participants = append(participants, p)
	}

	// connect all participants together
	Network(t, participants...)
	Bootstrap(t, participants...)

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

	// create the participants
	participantIDs := fixture.Hashes(t, n)
	participants := make([]*Participant, 0, len(participantIDs))
	for index, selfID := range participantIDs {
		log := zerolog.New(os.Stderr).With().
			Timestamp().
			Int("index", index).
			Hex("self", selfID[:]).
			Logger()
		p := NewParticipant(t,
			WithLog(log),
			WithSelf(selfID),
			WithParticipants(participantIDs),
			WithStop(
				AfterFinal(256, errFinished),
				AfterDelay(32*time.Second, errTimeout),
			),
		)
		participants = append(participants, p)
	}

	// connect all participants together
	Network(t, participants...)
	Bootstrap(t, participants...)

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

	// create the participants
	participantIDs := fixture.Hashes(t, n)
	participants := make([]*Participant, 0, len(participantIDs))
	for index, selfID := range participantIDs {
		log := zerolog.New(os.Stderr).With().
			Timestamp().
			Int("index", index).
			Hex("self", selfID[:]).
			Logger()
		p := NewParticipant(t,
			WithLog(log),
			WithSelf(selfID),
			WithParticipants(participantIDs),
			WithStop(
				AfterFinal(4, errFinished),
				AfterDelay(8*time.Second, errTimeout),
			),
		)
		participants = append(participants, p)
	}

	// connect all participants together
	Network(t, participants...)
	Bootstrap(t, participants...)

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
