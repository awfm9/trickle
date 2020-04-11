package integration

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"

	"github.com/alvalor/consensus"
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model"
)

type Participant struct {

	// participant data
	selfID         model.Hash
	participantIDs []model.Hash
	proposals      chan *message.Proposal
	votes          chan *message.Vote

	// processor data
	round      uint64
	blockDB    map[model.Hash]*model.Block
	proposalDB map[model.Hash]*message.Proposal
	voteDB     map[model.Hash](map[model.Hash]*message.Vote)

	// component mocks
	state  *mocks.State
	net    *mocks.Network
	build  *mocks.Builder
	sign   *mocks.Signer
	verify *mocks.Verifier
	buffer *mocks.Buffer

	// participant processor
	pro *consensus.Processor

	// test-related dependencies
	log  zerolog.Logger
	wg   sync.WaitGroup
	stop Condition
}

func NewParticipant(t require.TestingT, selfID model.Hash, stop Condition) *Participant {

	p := Participant{
		selfID:         selfID,
		participantIDs: []model.Hash{selfID},
		proposals:      make(chan *message.Proposal, 1024),
		votes:          make(chan *message.Vote, 1024),

		round:      0,
		blockDB:    make(map[model.Hash]*model.Block),
		proposalDB: make(map[model.Hash]*message.Proposal),
		voteDB:     make(map[model.Hash](map[model.Hash]*message.Vote)),

		net:    &mocks.Network{},
		state:  &mocks.State{},
		sign:   &mocks.Signer{},
		verify: &mocks.Verifier{},
		buffer: &mocks.Buffer{},
		build:  &mocks.Builder{},

		wg:   sync.WaitGroup{},
		stop: stop,
		log:  zerolog.New(os.Stderr).With().Hex("self", selfID[:]).Logger(),
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
			p.log.Info().Msg("broadcasting proposal")
			p.proposals <- proposal
			return nil
		},
	)
	p.net.On("Transmit", mock.Anything, mock.Anything).Return(
		func(vote *message.Vote, recipientID model.Hash) error {
			if recipientID == p.selfID {
				p.log.Info().Msg("transmitting vote")
				p.votes <- vote
			}
			return nil
		},
	)

	// program random builder behaviour
	p.build.On("PayloadHash").Return(
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
		func(block *model.Block) *message.Proposal {
			block.SignerID = p.selfID
			proposal := message.Proposal{
				Block:     block,
				Signature: nil,
			}
			return &proposal
		},
		nil,
	)
	p.sign.On("Vote", mock.Anything).Return(
		func(block *model.Block) *message.Vote {
			vote := message.Vote{
				Height:    block.Height,
				BlockID:   block.ID(),
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

	// program single-block buffer behaviour
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
			tally, hasBlock := p.voteDB[vote.BlockID]
			if !hasBlock {
				tally = make(map[model.Hash]*message.Vote)
				p.voteDB[vote.BlockID] = tally
			}
			_, hasVote := tally[vote.SignerID]
			if hasVote {
				return false
			}
			tally[vote.SignerID] = vote
			return true
		},
		func(vote *message.Vote) error {
			tally, hasBlock := p.voteDB[vote.BlockID]
			if !hasBlock {
				return nil
			}
			duplicate, hasVote := tally[vote.SignerID]
			if !hasVote {
				return nil
			}
			if vote.BlockID == duplicate.BlockID {
				return nil
			}
			return consensus.DoubleVote{First: duplicate, Second: vote}
		},
	)
	p.buffer.On("Votes", mock.Anything).Return(
		func(blockID model.Hash) []*message.Vote {
			votes := make([]*message.Vote, 0, len(p.votes))
			tally, tallied := p.voteDB[blockID]
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
			for blockID, proposal := range p.proposalDB {
				if proposal.Height <= height {
					delete(p.voteDB, blockID)
					delete(p.blockDB, blockID)
				}
			}
			for blockID, tally := range p.voteDB {
				for _, vote := range tally {
					if vote.Height <= height {
						delete(p.voteDB, blockID)
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

	for {

		// TODO: implement acceptable error filters once error types are
		// returned by the processor

		// check stop condition first
		if p.stop(p) {
			return nil
		}

		// check for first priority messages (proposals)
		select {
		case proposal := <-p.proposals:
			err := p.pro.OnProposal(proposal)
			if err != nil {
				p.log.Error().Err(err).Msg("could not process proposal")
			} else {
				p.log.Info().Msg("proposal processed")
			}
		default:
		}

		// check for second priority messages (votes)
		select {
		case proposal := <-p.proposals:
			err := p.pro.OnProposal(proposal)
			if err != nil {
				p.log.Error().Err(err).Msg("could not process proposal")
			} else {
				p.log.Info().Msg("proposal processed")
			}
		case vote := <-p.votes:
			p.log.Info().Msg("processing vote")
			err := p.pro.OnVote(vote)
			if err != nil {
				p.log.Error().Err(err).Msg("could not process vote")
			} else {
				p.log.Info().Msg("vote processed")
			}
		}
	}
}
