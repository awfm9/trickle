package integration

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/alvalor/consensus"
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model"
	"github.com/alvalor/consensus/test/fixture"
)

type Participant struct {

	// participant configuration
	log            zerolog.Logger
	selfID         model.Hash
	participantIDs []model.Hash
	stop           []Condition
	ignore         []error

	// processor data
	round      uint64
	vertexDB   map[model.Hash]*model.Vertex
	proposalDB map[model.Hash]*message.Proposal
	voteDB     map[model.Hash](map[model.Hash]*message.Vote)
	proposalQ  chan *message.Proposal
	voteQ      chan *message.Vote

	// processor dependencies
	state  *mocks.State
	net    *mocks.Network
	build  *mocks.Builder
	sign   *mocks.Signer
	verify *mocks.Verifier
	buffer *mocks.Buffer

	// processor instance
	pro *consensus.Processor
}

func NewParticipant(t require.TestingT, options ...Option) *Participant {

	// initialize the default values for the participant
	selfID := fixture.Hash(t)
	p := Participant{

		log:            zerolog.New(ioutil.Discard),
		selfID:         selfID,
		participantIDs: []model.Hash{selfID},
		round:          0,
		stop:           []Condition{AfterRound(10, errFinished)},
		ignore:         []error{consensus.ObsoleteProposal{}, consensus.ObsoleteVote{}},

		vertexDB:   make(map[model.Hash]*model.Vertex),
		proposalDB: make(map[model.Hash]*message.Proposal),
		voteDB:     make(map[model.Hash](map[model.Hash]*message.Vote)),
		proposalQ:  make(chan *message.Proposal, 1024),
		voteQ:      make(chan *message.Vote, 1024),

		net:    &mocks.Network{},
		state:  &mocks.State{},
		sign:   &mocks.Signer{},
		verify: &mocks.Verifier{},
		buffer: &mocks.Buffer{},
		build:  &mocks.Builder{},
	}

	// apply the options
	for _, option := range options {
		option(&p)
	}

	// program round-robin state behaviour
	p.state.On("Round").Return(
		func() uint64 {
			return p.round
		},
		nil,
	)
	p.state.On("Set", mock.Anything).Return(
		func(height uint64) error {
			if height <= p.round {
				return fmt.Errorf("must set higher round (set: %d, round: %d)", height, p.round)
			}
			p.round = height
			return nil
		},
	)
	p.state.On("Leader", mock.Anything).Return(
		func(height uint64) model.Hash {
			index := int(height % uint64(len(p.participantIDs)))
			leader := p.participantIDs[index]
			return leader
		},
		nil,
	)
	p.state.On("Threshold", mock.Anything).Return(
		func() uint {
			threshold := uint(len(p.participantIDs) * 2 / 3)
			return threshold
		},
		nil,
	)

	// program loopback network mock behaviour
	p.net.On("Broadcast", mock.Anything).Return(
		func(proposal *message.Proposal) error {
			p.proposalQ <- proposal
			return nil
		},
	)
	p.net.On("Transmit", mock.Anything, mock.Anything).Return(
		func(vote *message.Vote, recipientID model.Hash) error {
			if recipientID != p.selfID {
				return fmt.Errorf("invalid recipient (%x)", recipientID)
			}
			p.voteQ <- vote
			return nil
		},
	)

	// program random builder behaviour
	p.build.On("Arc").Return(
		func() model.Hash {
			seed := make([]byte, 128)
			n, err := rand.Read(seed)
			require.NoError(t, err, "could not read seed")
			require.Equal(t, len(seed), n, "could not read full seed")
			hash := sha3.Sum256(seed)
			return model.Hash(hash)
		},
		nil,
	)

	// program no-signature signer behaviour
	p.sign.On("Self").Return(
		func() model.Hash {
			return p.selfID
		},
		nil,
	)
	p.sign.On("Proposal", mock.Anything).Return(
		func(vertex *model.Vertex) *message.Proposal {
			vertex.SignerID = p.selfID
			proposal := message.Proposal{
				Vertex:    vertex,
				Signature: nil,
			}
			return &proposal
		},
		nil,
	)
	p.sign.On("Vote", mock.Anything).Return(
		func(vertex *model.Vertex) *message.Vote {
			vote := message.Vote{
				Height:    vertex.Height,
				VertexID:  vertex.ID(),
				SignerID:  p.selfID,
				Signature: nil,
			}
			return &vote
		},
		nil,
	)

	// program always-valid verifier behaviour
	p.verify.On("Proposal", mock.Anything).Return(nil)
	p.verify.On("Vote", mock.Anything).Return(nil)

	// program single-vertex buffer behaviour
	p.buffer.On("Proposal", mock.Anything).Return(
		func(proposal *message.Proposal) bool {
			_, hasProposal := p.proposalDB[proposal.ID()]
			return !hasProposal
		},
		func(proposal *message.Proposal) error {
			for _, duplicate := range p.proposalDB {
				if proposal.Height == duplicate.Height && proposal.SignerID == duplicate.SignerID {
					return consensus.DoubleProposal{First: duplicate, Second: proposal}
				}
			}
			return nil
		},
	)
	p.buffer.On("Vote", mock.Anything).Return(
		func(vote *message.Vote) bool {
			tally, hasVertex := p.voteDB[vote.VertexID]
			if !hasVertex {
				tally = make(map[model.Hash]*message.Vote)
				p.voteDB[vote.VertexID] = tally
			}
			_, hasVote := tally[vote.SignerID]
			if hasVote {
				return false
			}
			tally[vote.SignerID] = vote
			return true
		},
		func(vote *message.Vote) error {
			tally, hasVertex := p.voteDB[vote.VertexID]
			if !hasVertex {
				return nil
			}
			duplicate, hasVote := tally[vote.SignerID]
			if !hasVote {
				return nil
			}
			if vote.VertexID == duplicate.VertexID {
				return nil
			}
			return consensus.DoubleVote{First: duplicate, Second: vote}
		},
	)
	p.buffer.On("Votes", mock.Anything).Return(
		func(VertexID model.Hash) []*message.Vote {
			var votes []*message.Vote
			tally, tallied := p.voteDB[VertexID]
			if tallied {
				for _, vote := range tally {
					votes = append(votes, vote)
				}
			}
			return votes
		},
		nil,
	)
	p.buffer.On("Clear", mock.Anything).Return(
		func(height uint64) error {
			for VertexID, proposal := range p.proposalDB {
				if proposal.Height <= height {
					delete(p.voteDB, VertexID)
					delete(p.vertexDB, VertexID)
				}
			}
			for VertexID, tally := range p.voteDB {
				for _, vote := range tally {
					if vote.Height <= height {
						delete(p.voteDB, VertexID)
					}
					break
				}
			}
			return nil
		},
	)

	p.pro = consensus.NewProcessor(p.state, p.net, p.build, p.sign, p.verify, p.buffer)

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
				return err
			}
		}

		// wait for a message to process
		var err error
		var entity string
		select {
		case proposal := <-p.proposalQ:
			entity = "proposal"
			err = p.pro.OnProposal(proposal)
		case vote := <-p.voteQ:
			entity = "vote"
			err = p.pro.OnVote(vote)
		case <-time.After(100 * time.Millisecond):
			continue
		}

		// check if we should log a message
		if err == nil {
			p.log.Info().Str("entity", entity).Msg("message processed")
			continue
		}
		if ignore(err) {
			p.log.Warn().Str("entity", entity).Msg("could not process message")
			continue
		}

		return fmt.Errorf("could not process message: %w", err)
	}
}
