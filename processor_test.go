package consensus

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/awfm/consensus/mocks"
	"github.com/awfm/consensus/model/base"
	"github.com/awfm/consensus/model/fixture"
	"github.com/awfm/consensus/model/message"
	"github.com/awfm/consensus/model/signal"
)

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

type ProcessorSuite struct {
	suite.Suite

	// parameters for signer
	self base.Hash

	// parameters for graph mock
	final   *base.Vertex
	tip     *base.Vertex
	staleID base.Hash

	// parameters for strategy mock
	leaderID    base.Hash
	collectorID base.Hash

	// mocked dependencies
	net    *mocks.Network
	graph  *mocks.Graph
	build  *mocks.Builder
	strat  *mocks.Strategy
	sign   *mocks.Signer
	verify *mocks.Verifier
	cache  *mocks.Cache

	// processor under test
	pro *Processor
}

func (ps *ProcessorSuite) SetupTest() {

	// parameters for signer
	ps.self = fixture.Hash(ps.T())

	// parameters for the graph mock
	ps.final = fixture.Vertex(ps.T())
	between := fixture.Vertex(ps.T(), fixture.WithParent(ps.final))
	ps.tip = fixture.Vertex(ps.T(), fixture.WithParent(between))
	ps.staleID = fixture.Hash(ps.T())

	// parameters for the strategy mock
	ps.leaderID = fixture.Hash(ps.T())
	ps.collectorID = fixture.Hash(ps.T())

	// initialize the mocked dependencies
	ps.net = &mocks.Network{}
	ps.graph = &mocks.Graph{}
	ps.build = &mocks.Builder{}
	ps.strat = &mocks.Strategy{}
	ps.sign = &mocks.Signer{}
	ps.verify = &mocks.Verifier{}
	ps.cache = &mocks.Cache{}

	// program the signer mock
	ps.sign.On("Self").Return(
		func() base.Hash {
			return ps.self
		},
		nil,
	).Maybe()
	ps.sign.On("Vote", mock.Anything).Return(
		func(candidate *base.Vertex) *message.Vote {
			vote := message.Vote{
				Height:      candidate.Height,
				CandidateID: candidate.ID(),
				SignerID:    ps.self,
				Signature:   nil,
			}
			return &vote
		},
		nil,
	).Maybe()

	// program the graph mock
	ps.graph.On("Final").Return(
		func() *base.Vertex {
			return ps.final
		},
		nil,
	).Maybe()
	ps.graph.On("Tip").Return(
		func() *base.Vertex {
			return ps.tip
		},
		nil,
	).Maybe()
	ps.graph.On("Contains", mock.Anything).Return(
		func(vertexID base.Hash) bool {
			return vertexID == ps.staleID
		},
		nil,
	).Maybe()

	// program strategy mock
	ps.strat.On("Leader", mock.Anything).Return(
		func(height uint64) base.Hash {
			return ps.leaderID
		},
		nil,
	).Maybe()
	ps.strat.On("Collector", mock.Anything).Return(
		func(height uint64) base.Hash {
			return ps.collectorID
		},
		nil,
	).Maybe()

	// program verify mock
	ps.verify.On("Proposal", mock.Anything).Return(nil).Maybe()

	// initialize the processor under test
	ps.pro = NewProcessor(ps.net, ps.graph, ps.build, ps.strat, ps.sign, ps.verify, ps.cache)
}

func (ps *ProcessorSuite) TestBootstrap() {

	// make sure bootstrapping fails with non-zero genesis height
	ps.tip.Height = 1
	err := ps.pro.Bootstrap()
	require.Error(ps.T(), err, "should fail bootstrapping with height one")
	ps.net.AssertNumberOfCalls(ps.T(), "Transmit", 0)

	// make sure we transmit the right vote on successful bootstrap
	ps.net.On("Transmit", mock.Anything, mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			vote := args.Get(0).(*message.Vote)
			collectorID := args.Get(1).(base.Hash)
			require.Equal(ps.T(), ps.tip.Height, vote.Height, "should send vote for tip height")
			require.Equal(ps.T(), ps.tip.ID(), vote.CandidateID, "should send vote for tip vertex")
			require.Equal(ps.T(), ps.self, vote.SignerID, "should send vote by self")
			require.Equal(ps.T(), ps.collectorID, collectorID, "should send vote to collector")
		},
	)

	// make sure valid bootstrap passes
	ps.tip.Height = 0
	err = ps.pro.Bootstrap()
	require.NoError(ps.T(), err, "should successfully bootstrap with height zero")
	ps.net.AssertExpectations(ps.T())
}

func (ps *ProcessorSuite) TestConfirmParent() {

	// create parent, candidate and proposal
	parent := fixture.Vertex(ps.T())
	candidate := fixture.Vertex(ps.T(), fixture.WithParent(parent))
	proposal := fixture.Proposal(ps.T(), fixture.WithCandidate(candidate))

	// check that the quorum is checked correctly
	ps.verify.On("Quorum", mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			verified := args.Get(0).(*message.Proposal)
			require.Equal(ps.T(), proposal, verified, "should verify proposal quorum")
		},
	)

	// make sure the parent is confirmed correctly
	ps.graph.On("Confirm", mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			confirmedID := args.Get(0).(base.Hash)
			require.Equal(ps.T(), parent.ID(), confirmedID, "should confirm proposal parent")
		},
	)

	// make sure the cache is cleared up to parent height
	ps.cache.On("Clear", mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			cleared := args.Get(0).(uint64)
			require.Equal(ps.T(), parent.Height, cleared, "should clear up to parent height")
		},
	)

	// execute the function
	err := ps.pro.confirmParent(proposal)
	require.NoError(ps.T(), err, "should successfully confirm parent")
	ps.verify.AssertExpectations(ps.T())
	ps.graph.AssertExpectations(ps.T())
	ps.cache.AssertExpectations(ps.T())
}

func (ps *ProcessorSuite) TestApplyCandidate() {

	// create candidate and proposal
	candidate := fixture.Vertex(ps.T(), fixture.WithProposer(ps.leaderID), fixture.WithParent(ps.tip))
	proposal := fixture.Proposal(ps.T(), fixture.WithCandidate(candidate))

	// make sure we get a stale error if graph already contains candidate
	ps.staleID = candidate.ID()
	err := ps.pro.applyCandidate(proposal)
	require.Error(ps.T(), err, "should not pass stale candidate")
	require.True(ps.T(), errors.As(err, &signal.StaleProposal{}), "should have stale proposal error")
	ps.graph.AssertNumberOfCalls(ps.T(), "Extend", 0)
	ps.staleID = base.ZeroHash

	// make sure we get an invalid proposer error with wrong proposer
	tempID := ps.leaderID
	ps.leaderID = fixture.Hash(ps.T())
	err = ps.pro.applyCandidate(proposal)
	require.Error(ps.T(), err, "should not pass with wrong proposer")
	require.True(ps.T(), errors.As(err, &signal.InvalidProposer{}), "should have invalid proposer error")
	ps.graph.AssertNumberOfCalls(ps.T(), "Extend", 0)
	ps.leaderID = tempID

	// make sure that a height at final leads to conflicting error
	height := candidate.Height
	candidate.Height = ps.final.Height
	err = ps.pro.applyCandidate(proposal)
	require.Error(ps.T(), err, "should not pass conflicting candidate")
	require.True(ps.T(), errors.As(err, &signal.ConflictingProposal{}), "should have conflicting proposal error")
	ps.graph.AssertNumberOfCalls(ps.T(), "Extend", 0)
	candidate.Height = height

	// make sure that a height below tip leads to conflicting error
	candidate.Height = ps.tip.Height - 1
	err = ps.pro.applyCandidate(proposal)
	require.Error(ps.T(), err, "should not pass obsolete candidate")
	require.True(ps.T(), errors.As(err, &signal.ObsoleteProposal{}), "should have conflicting proposal error")
	ps.graph.AssertNumberOfCalls(ps.T(), "Extend", 0)
	candidate.Height = height

	// make sure the graph is extended with the proposal vertex
	ps.graph.On("Extend", mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			extended := args.Get(0).(*base.Vertex)
			require.Equal(ps.T(), candidate, extended, "should extend with proposal candidate")
		},
	)

	// make sure the proposal is stored in the cache for double detection
	ps.cache.On("Proposal", mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			cached := args.Get(0).(*message.Proposal)
			require.Equal(ps.T(), proposal, cached, "should cache the applied proposal")
		},
	)

	// make sure we pass with valid proposal
	err = ps.pro.applyCandidate(proposal)
	require.NoError(ps.T(), err, "should pass valid proposal")
	ps.graph.AssertExpectations(ps.T())
	ps.cache.AssertExpectations(ps.T())
}
