package consensus

import (
	"crypto/rand"
	"testing"

	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/test/fixture"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/sha3"
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
	pro    *Processor
}

func (ps *ProcessorSuite) SetupTest() {

	// placeholder for random seed
	seed := make([]byte, 256)

	// generate identity
	_, _ = rand.Read(seed)
	self := sha3.Sum256(seed)

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

	// set up the processor
	ps.pro = NewProcessor(ps.db, ps.net, ps.state, ps.sign, ps.verify, self)
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

	// set our own ID
	ps.pro.selfID = selfID

	// program the state
	ps.state.On("Height").Return(block.Height)
	ps.state.On("Leader", block.Height).Return(leaderID)
	ps.state.On("Leader", block.Height+1).Return(collectorID)
	ps.state.On("Set", block.Height).Return()

	// program the verifier
	ps.verify.On("Proposal", proposal).Return(true, nil)

	// program the database
	ps.db.On("Block", block.QC.BlockID).Return(parent, nil)

	// program the signer
	ps.sign.On("Vote", block).Return(vote, nil)

	// program the network
	ps.net.On("Transmit", vote, collectorID).Return(nil)

	// check everything happens as desired
	err := ps.pro.OnProposal(proposal)
	require.NoError(ps.T(), err, "proposal should be valid")
}

func (ps *ProcessorSuite) TestProcessorOnVote() {

}
