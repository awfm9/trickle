package integration

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"

	"github.com/alvalor/consensus"
	"github.com/alvalor/consensus/cache"
	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
	"github.com/alvalor/consensus/model/signal"
	"github.com/alvalor/consensus/strategy"
	"github.com/alvalor/consensus/test/fixture"
)

type Participant struct {

	// participant configuration
	log            zerolog.Logger
	selfID         base.Hash
	participantIDs []base.Hash
	genesisID      base.Hash
	stop           []Condition
	ignore         []error

	// state data
	final         *base.Vertex
	confirmations map[base.Hash]uint
	vertexDB      map[base.Hash]*base.Vertex
	proposalDB    map[base.Hash]*message.Proposal
	voteCache     map[base.Hash](map[base.Hash]*message.Vote)
	proposalQ     chan *message.Proposal
	voteQ         chan *message.Vote

	// real dependencies
	strat consensus.Strategy
	cache consensus.Cache

	// dependency mocks
	net    *mocks.Network
	graph  *mocks.Graph
	build  *mocks.Builder
	sign   *mocks.Signer
	verify *mocks.Verifier

	// processor instance
	pro *consensus.Processor
}

func NewParticipant(t testing.TB, options ...Option) *Participant {

	// configure time function for zerolog
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }

	// initialize the default values for the participant
	selfID := fixture.Hash(t)
	p := Participant{

		log:            zerolog.New(ioutil.Discard),
		selfID:         selfID,
		participantIDs: []base.Hash{selfID},
		genesisID:      base.ZeroHash,
		stop:           []Condition{AfterDelay(time.Second, errFinished)},
		ignore: []error{
			signal.StaleProposal{},
			signal.StaleVote{},
			signal.ObsoleteProposal{},
			signal.ObsoleteVote{},
		},

		final:         nil,
		confirmations: make(map[base.Hash]uint),
		vertexDB:      make(map[base.Hash]*base.Vertex),
		proposalDB:    make(map[base.Hash]*message.Proposal),
		voteCache:     make(map[base.Hash](map[base.Hash]*message.Vote)),
		proposalQ:     make(chan *message.Proposal, 1024),
		voteQ:         make(chan *message.Vote, 1024),

		net:    &mocks.Network{},
		graph:  &mocks.Graph{},
		build:  &mocks.Builder{},
		sign:   &mocks.Signer{},
		verify: &mocks.Verifier{},
	}

	// apply the options
	for _, option := range options {
		option(&p)
	}

	// initialize the real dependencies
	p.strat = strategy.NewNaive(p.participantIDs)
	p.cache = cache.NewVolatile()

	// program loopback network mock behaviour
	p.net.On("Broadcast", mock.Anything).Return(
		func(proposal *message.Proposal) error {
			p.proposalQ <- proposal
			candidateID := proposal.Candidate.ID()
			p.log.Debug().
				Str("message", "proposal").
				Uint64("height", proposal.Candidate.Height).
				Hex("candidate", candidateID[:]).
				Hex("parent", proposal.Candidate.ParentID[:]).
				Hex("arc", proposal.Candidate.ArcID[:]).
				Hex("proposer", proposal.Candidate.ProposerID[:]).
				Msg("proposal looped")
			return nil
		},
	)
	p.net.On("Transmit", mock.Anything, mock.Anything).Return(nil)

	// program simple graph behaviour
	p.graph.On("Extend", mock.Anything).Return(
		func(vertex *base.Vertex) error {
			if p.final != nil && vertex.Height <= p.final.Height {
				return fmt.Errorf("vertex conflicts with finalized state")
			}
			p.vertexDB[vertex.ID()] = vertex
			return nil
		},
	)
	p.graph.On("Confirm", mock.Anything).Return(
		func(vertexID base.Hash) error {
			vertex, hasVertex := p.vertexDB[vertexID]
			if !hasVertex {
				return fmt.Errorf("could not find vertex (%x)", vertexID)
			}
			p.confirmations[vertexID]++
			if p.confirmations[vertexID] < 3 {
				return nil
			}
			if p.final == nil {
				p.final = vertex
				return nil
			}
			if vertex.Height < p.final.Height {
				return nil
			}
			if vertex.Height > p.final.Height {
				p.final = vertex
				return nil
			}
			if p.confirmations[vertexID] > p.confirmations[p.final.ID()] {
				p.final = vertex
				return nil
			}
			return nil
		},
	)
	p.graph.On("Final").Return(
		func() *base.Vertex {
			return p.final
		},
		func() bool {
			return p.final != nil
		},
	)

	// program random builder behaviour
	p.build.On("Arc").Return(
		func() base.Hash {
			return fixture.Hash(t)
		},
		nil,
	)

	// program no-signature signer behaviour
	p.sign.On("Self").Return(
		func() base.Hash {
			return p.selfID
		},
		nil,
	)
	p.sign.On("Proposal", mock.Anything).Return(
		func(candidate *base.Vertex) *message.Proposal {
			candidate.ProposerID = p.selfID
			proposal := message.Proposal{
				Candidate: candidate,
				Signature: nil,
			}
			return &proposal
		},
		nil,
	)
	p.sign.On("Vote", mock.Anything).Return(
		func(candidate *base.Vertex) *message.Vote {
			vote := message.Vote{
				Height:      candidate.Height,
				CandidateID: candidate.ID(),
				SignerID:    p.selfID,
				Signature:   nil,
			}
			return &vote
		},
		nil,
	)

	// program always-valid verifier behaviour
	p.verify.On("Proposal", mock.Anything).Return(nil)
	p.verify.On("Vote", mock.Anything).Return(nil)

	// inject dependencies into processor
	p.pro = consensus.NewProcessor(p.net, p.graph, p.build, p.strat, p.sign, p.verify, p.cache)

	return &p
}

func (p *Participant) Run() error {

	// create the ignore function to easily check errors
	ignore := Ignore(Combine(p.ignore...))

	for {

		// check stop condition first
		for _, condition := range p.stop {
			err := condition(p)
			if err != nil {
				return fmt.Errorf("encountered stop condition: %w", err)
			}
		}

		// set up logging entry and error
		var err error

		// wait for a message to process
		select {
		case proposal := <-p.proposalQ:
			candidateID := proposal.Candidate.ID()
			p.log.Debug().
				Uint64("height", proposal.Candidate.Height).
				Hex("candidate", candidateID[:]).
				Hex("parent", proposal.Candidate.ParentID[:]).
				Hex("arc", proposal.Candidate.ArcID[:]).
				Hex("proposer", proposal.Candidate.ProposerID[:]).
				Msg("proposal received")
			err = p.pro.OnProposal(proposal)
		case vote := <-p.voteQ:
			p.log.Debug().
				Uint64("height", vote.Height).
				Hex("candidate", vote.CandidateID[:]).
				Hex("voter", vote.SignerID[:]).
				Msg("vote received")
			err = p.pro.OnVote(vote)
		case <-time.After(100 * time.Millisecond):
			continue
		}

		// check if we should log a message
		if err == nil || ignore(err) {
			continue
		}

		return fmt.Errorf("encountered critical error: %w", err)
	}
}
