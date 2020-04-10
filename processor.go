package consensus

import (
	"fmt"

	"github.com/alvalor/consensus/event"
	"github.com/alvalor/consensus/model"
)

type Processor struct {
	net    Network
	state  State
	signer Signer
	self   model.Hash
}

func (pro *Processor) OnTransition(trans *event.Transition) error {

	// set our state to the new height
	pro.state.Set(trans.Block.Height)

	// TODO: check if we have cached proposals for this state

	// vote on the new proposal
	// NOTE: if we are the next leader, the network will short-circuit the vote
	// here, and the `OnVote` call will trigger us sending the proposal if we
	// can already assemble enough signatures for the previous one
	vote := pro.signer.Vote(trans.Block)
	collectorID := pro.state.Collector(vote.Height)
	err := pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnVote(vote *event.Vote) error {

	// check if the vote is outdated
	height := pro.state.Height()
	if vote.Height < height {
		return fmt.Errorf("outdated vote height (vote: %d, round: %d)", vote.Height, height)
	}

	// TODO: validate vote - signer & signature

	// check if we are collector for the round
	collector := pro.state.Collector(height)
	if collector != pro.self {
		return fmt.Errorf("invalid vote recipient (collector: %x, self: %x)", collector, pro.self)
	}

	// TODO: add the vote to the accumulator

	// check if the vote is future
	if vote.Height > height {
		return nil
	}

	// TODO: see if we can build our proposal
	proposal := &event.Proposal{}

	// broadcast the proposal to the network
	// NOTE: the network module should short-circuit one copy of this message to
	// ourselves, which will lead to the state transition to the next height
	err := pro.net.Broadcast(proposal)
	if err != nil {
		return fmt.Errorf("could not broadcast proposal: %w", err)
	}

	return nil
}

func (pro *Processor) OnProposal(proposal *event.Proposal) error {

	// check if the proposal is outdated
	height := pro.state.Height()
	if proposal.Block.Height < height {
		return fmt.Errorf("outdated proposal height (proposal: %d, round: %d)", proposal.Block.Height, height)
	}

	// check if the proposal is by correct leader
	leaderID := pro.state.Leader(height)
	if proposal.Block.LeaderID != leaderID {
		return fmt.Errorf("wrong proposal leader (proposal: %x, leader: %x)", proposal.Block.LeaderID, leaderID)
	}

	// TODO: validate proposal - signer & signature

	// check if the proposal is future
	if proposal.Block.Height > height {
		// TODO: buffer for later
		return nil
	}

	// TODO: verify proposal - state extension

	// set our state to the new height
	pro.state.Set(proposal.Block.Height)

	// TODO: check if we have cached proposals for new height

	// check if we were the proposer
	if proposal.Block.LeaderID == pro.self {
		return nil
	}

	// vote on the new proposal
	// NOTE: if we are the next leader, the network will short-circuit the vote
	// here, and the `OnVote` call will trigger us sending the proposal if we
	// can already assemble enough signatures for the previous one
	vote := pro.signer.Vote(proposal.Block)
	collectorID := pro.state.Collector(vote.Height)
	err := pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit vote: %w", err)
	}

	return nil
}
