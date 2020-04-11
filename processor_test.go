package consensus

import (
	"testing"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model"
	"github.com/alvalor/consensus/test/fixture"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

type ProcessorSuite struct {
	suite.Suite
	db     *mocks.Database
	net    *mocks.Network
	state  *mocks.State
	sign   *mocks.Signer
	verify *mocks.Verifier
	buf    *mocks.Buffer
	build  *mocks.Builder
	pro    *Processor
}

func (ps *ProcessorSuite) SetupTest() {

	// set up database mock
	ps.db = &mocks.Database{}

	// set up network mock
	ps.net = &mocks.Network{}

	// set up state mock
	ps.state = &mocks.State{}

	// set up signer mock
	ps.sign = &mocks.Signer{}

	// set up verify mock
	ps.verify = &mocks.Verifier{}

	// set up buffer mock
	ps.buf = &mocks.Buffer{}

	// set up builder mock
	ps.build = &mocks.Builder{}

	// set up the processor
	ps.pro = NewProcessor(ps.db, ps.net, ps.state, ps.sign, ps.verify, ps.buf, ps.build)
}

func (ps *ProcessorSuite) TestProcessorOnProposal() {

	// define the actors for our experiment
	selfID := fixture.Hash(ps.T())
	leaderID := fixture.Hash(ps.T())
	collectorID := fixture.Hash(ps.T())

	// create the parent block for current candidate
	parent := fixture.Block(ps.T())

	// create the block to be voted upon in current round
	block := fixture.Block(ps.T(),
		fixture.WithParent(parent),
		fixture.WithProposer(leaderID),
	)

	// create the proposal proposing the current block
	proposal := fixture.Proposal(ps.T(),
		fixture.WithBlock(block),
	)

	// create our potential vote for the proposal
	vote := fixture.Vote(ps.T(),
		fixture.WithCandidate(block),
		fixture.WithVoter(selfID),
	)

	// program the state
	ps.state.On("Round").Return(block.Height)
	ps.state.On("Leader", block.Height).Return(leaderID)
	ps.state.On("Leader", block.Height+1).Return(collectorID)
	ps.state.On("Set", block.Height).Return()

	// program the verifier
	ps.verify.On("Proposal", proposal).Return(true, nil)

	// program the database
	ps.db.On("Block", block.QC.BlockID).Return(parent, nil)

	// program the signer
	ps.sign.On("Self").Return(selfID)
	ps.sign.On("Vote", block).Return(vote, nil)

	// program the network
	ps.net.On("Transmit", vote, collectorID).Return(nil)

	// check everything happens as desired
	err := ps.pro.OnProposal(proposal)
	require.NoError(ps.T(), err, "valid proposal should pass")
}

func (ps *ProcessorSuite) TestProcessorOnVote() {

	// define the actors for our experiment
	selfID := fixture.Hash(ps.T())
	voterID := fixture.Hash(ps.T())

	// set up the candidate block voted upon
	candidate := fixture.Block(ps.T())

	// generate random payload hash for next block
	payloadHash := fixture.Hash(ps.T())

	// create the vote we will receive
	vote := fixture.Vote(ps.T(),
		fixture.WithCandidate(candidate),
		fixture.WithVoter(voterID),
	)

	// program the database
	ps.db.On("Block", vote.BlockID).Return(candidate, nil)

	// program the state
	ps.state.On("Round").Return(candidate.Height)
	ps.state.On("Leader", candidate.Height+1).Return(selfID)
	ps.state.On("Threshold", candidate.Height).Return(uint(0))

	// program the verifier
	ps.verify.On("Vote", vote).Return(true, nil)

	// program the buffer
	ps.buf.On("Tally", vote).Return(nil)
	ps.buf.On("Votes", vote.BlockID).Return([]*message.Vote{vote})

	// program the builder
	ps.build.On("PayloadHash").Return(payloadHash, nil)

	// program the signer
	ps.sign.On("Self").Return(selfID)
	ps.sign.On("Proposal", mock.Anything).Return(
		func(block *model.Block) *message.Proposal {
			proposal := message.Proposal{
				Block:     block,
				Signature: fixture.Sig(ps.T()),
			}
			return &proposal
		},
		nil,
	)

	// program the network
	ps.net.On("Broadcast", mock.Anything).Return(nil)

	// check everything happens as desired
	err := ps.pro.OnVote(vote)
	require.NoError(ps.T(), err, "valid vote should pass")
}
