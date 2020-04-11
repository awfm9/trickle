package integration

import (
	"crypto/rand"
	"fmt"

	"github.com/stretchr/testify/assert"
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
	queue          chan interface{}

	// processor data
	round  uint64
	blocks map[model.Hash]*model.Block
	votes  map[model.Hash]*message.Vote

	// component mocks
	db     *mocks.Database
	net    *mocks.Network
	state  *mocks.State
	sign   *mocks.Signer
	verify *mocks.Verifier
	buf    *mocks.Buffer
	build  *mocks.Builder

	// participant processor
	pro *consensus.Processor
}

func NewParticipant(t require.TestingT, selfID model.Hash) *Participant {

	p := Participant{
		selfID:         selfID,
		participantIDs: []model.Hash{selfID},
		queue:          make(chan interface{}, 1024),
		round:          0,
		blocks:         make(map[model.Hash]*model.Block),
		votes:          make(map[model.Hash]*message.Vote),
		db:             &mocks.Database{},
		net:            &mocks.Network{},
		state:          &mocks.State{},
		sign:           &mocks.Signer{},
		verify:         &mocks.Verifier{},
		buf:            &mocks.Buffer{},
		build:          &mocks.Builder{},
	}

	// program map database behaviour
	p.db.On("Store", mock.Anything).Return(
		func(block *model.Block) error {
			blockID := block.ID()
			_, exists := p.blocks[blockID]
			if exists {
				return fmt.Errorf("block already exists (%x)", blockID)
			}
			p.blocks[blockID] = block
			return nil
		},
	)
	p.db.On("Block", mock.Anything).Return(
		func(blockID model.Hash) *model.Block {
			block := p.blocks[blockID]
			return block
		},
		func(blockID model.Hash) error {
			_, exists := p.blocks[blockID]
			if !exists {
				return fmt.Errorf("unknown block (%x)", blockID)
			}
			return nil
		},
	)

	// program loopback network mock behaviour
	p.db.On("Broadcast", mock.Anything).Return(
		func(proposal *message.Proposal) error {
			p.queue <- proposal
			return nil
		},
	)
	p.db.On("Transmit", mock.Anything, mock.Anything).Return(
		func(vote *message.Vote, recipientID model.Hash) error {
			if recipientID == p.selfID {
				p.queue <- vote
			}
			return nil
		},
	)

	// program round-robin state behaviour
	p.state.On("Round").Return(
		func() uint64 {
			return p.round
		},
	)
	p.state.On("Set", mock.Anything).Return(
		func(height uint64) {
			p.round = height
		},
	)
	p.state.On("Leader", mock.Anything).Return(
		func(height uint64) model.Hash {
			index := int(height % uint64(len(p.participantIDs)))
			leader := p.participantIDs[index]
			return leader
		},
	)
	p.state.On("Threshold", mock.Anything).Return(
		func() uint {
			threshold := uint(len(p.participantIDs) * 2 / 3)
			return threshold
		},
	)

	// program no-signature signer behaviour
	p.sign.On("Self").Return(
		func() model.Hash {
			return p.selfID
		},
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
				BlockID:   block.ID(),
				SignerID:  p.selfID,
				Signature: nil,
			}
			return &vote
		},
		nil,
	)

	// program always-valid verifier behaviour
	p.verify.On("Proposal", mock.Anything).Return(true, nil)
	p.verify.On("Vote", mock.Anything).Return(true, nil)

	// program single-block buffer behaviour
	p.buf.On("Tally", mock.Anything).Return(
		func(vote *message.Vote) error {
			_, voted := p.votes[vote.SignerID]
			if voted {
				return fmt.Errorf("signer has already voted (%x)", vote.SignerID)
			}
			p.votes[vote.SignerID] = vote
			return nil
		},
	)
	p.buf.On("Votes", mock.Anything).Return(
		func(blockID model.Hash) []*message.Vote {
			votes := make([]*message.Vote, 0, len(p.votes))
			for _, vote := range p.votes {
				votes = append(votes, vote)
			}
			return votes
		},
	)
	p.buf.On("Clear", mock.Anything).Return(
		func(blockID model.Hash) {
			for signerID := range p.votes {
				delete(p.votes, signerID)
			}
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

	p.pro = consensus.NewProcessor(p.db, p.net, p.state, p.sign, p.verify, p.buf, p.build)

	go p.process(t)

	return &p
}

func (p *Participant) process(t require.TestingT) {

	// go through the queued messages
	for msg := range p.queue {

		// process the message
		switch m := msg.(type) {
		case *message.Proposal:
			err := p.pro.OnProposal(m)
			assert.NoError(t, err, "proposal should not error")
		case *message.Vote:
			err := p.pro.OnVote(m)
			assert.NoError(t, err, "vote should not error")
		default:
			panic(fmt.Sprintf("invalid message type (%T)", msg))
		}
	}
}
