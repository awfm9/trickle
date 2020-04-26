package consensus

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model"
	"github.com/alvalor/consensus/test/fixture"
)

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

type ProcessorSuite struct {
	suite.Suite

	// parameters for signer
	self model.Hash

	// parameters for graph mock
	final *model.Vertex
	tip   *model.Vertex

	// parameters for strategy mock
	leader    model.Hash
	collector model.Hash

	// mocked dependencies
	net       *mocks.Network
	graph     *mocks.Graph
	build     *mocks.Builder
	strat     *mocks.Strategy
	sign      *mocks.Signer
	verify    *mocks.Verifier
	proposals *mocks.ProposalCache
	votes     *mocks.VoteCache

	// processor under test
	pro *Processor
}

func (ps *ProcessorSuite) SetupTest() {

	// parameters for signer
	ps.self = fixture.Hash(ps.T())

	// parameters for the graph mock
	ps.final = fixture.Vertex(ps.T())
	ps.tip = fixture.Vertex(ps.T(), fixture.WithParent(ps.final))

	// parameters for the strategy mock
	ps.leader = fixture.Hash(ps.T())
	ps.collector = fixture.Hash(ps.T())

	// initialize the mocked dependencies
	ps.net = &mocks.Network{}
	ps.graph = &mocks.Graph{}
	ps.build = &mocks.Builder{}
	ps.strat = &mocks.Strategy{}
	ps.sign = &mocks.Signer{}
	ps.verify = &mocks.Verifier{}
	ps.proposals = &mocks.ProposalCache{}
	ps.votes = &mocks.VoteCache{}

	// program the signer mock
	ps.sign.On("Self").Return(
		func() model.Hash {
			return ps.self
		},
		nil,
	)
	ps.sign.On("Vote", mock.Anything).Return(
		func(vertex *model.Vertex) *message.Vote {
			vote := message.Vote{
				Height:    vertex.Height,
				VertexID:  vertex.ID(),
				SignerID:  ps.self,
				Signature: nil,
			}
			return &vote
		},
		nil,
	)

	// program the graph mock
	ps.graph.On("Final").Return(
		func() *model.Vertex {
			return ps.final
		},
		nil,
	)
	ps.graph.On("Tip").Return(
		func() *model.Vertex {
			return ps.tip
		},
		nil,
	)

	// program strategy mock
	ps.strat.On("Leader", mock.Anything).Return(
		func(height uint64) model.Hash {
			return ps.leader
		},
		nil,
	)
	ps.strat.On("Collector", mock.Anything).Return(
		func(height uint64) model.Hash {
			return ps.collector
		},
		nil,
	)

	// initialize the processor under test
	ps.pro = NewProcessor(ps.net, ps.graph, ps.build, ps.strat, ps.sign, ps.verify, ps.proposals, ps.votes)
}

func (ps *ProcessorSuite) TestBootstrap() {

	// 1) check that we send the correct vote to the correct recipient if we
	// bootstrap from a tip at height zero
	ps.tip.Height = 0
	ps.net.On("Transmit", mock.Anything, mock.Anything).Return(nil).Once().Run(
		func(args mock.Arguments) {
			vote := args.Get(0).(*message.Vote)
			collectorID := args.Get(1).(model.Hash)
			require.Equal(ps.T(), ps.tip.Height, vote.Height, "should send vote for tip height")
			require.Equal(ps.T(), ps.tip.ID(), vote.VertexID, "should send vote for tip vertex")
			require.Equal(ps.T(), ps.self, vote.SignerID, "should send vote by self")
			require.Equal(ps.T(), ps.collector, collectorID, "should send vote to collector")
		},
	)
	err := ps.pro.Bootstrap()
	require.NoError(ps.T(), err, "should bootstrap with height zero")

	// 2) check that we error if we try to bootstrap with a graph that has a tip
	// higher than height zero
	ps.tip.Height = 1
	err = ps.pro.Bootstrap()
	require.Error(ps.T(), err, "should not bootstrap with height one")
}
