// Consensus is a general purpose event-driven BFT consensus harness.
// Copyright (C) 2020 Max Wolter

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package consensus

import (
	"github.com/awfm/rich"

	"github.com/awfm/consensus/model/base"
	"github.com/awfm/consensus/model/message"
	"github.com/awfm/consensus/model/signal"
)

type Processor struct {
	net    Network
	graph  Graph
	build  Builder
	strat  Strategy
	sign   Signer
	verify Verifier
	cache  Cache
	loop   Looper
}

func NewProcessor(net Network, graph Graph, build Builder, strat Strategy, sign Signer, verify Verifier, cache Cache) *Processor {

	pro := Processor{
		net:    net,
		graph:  graph,
		build:  build,
		strat:  strat,
		sign:   sign,
		verify: verify,
		cache:  cache,
	}

	return &pro
}

func (pro *Processor) Bootstrap() error {

	// get current tip of the state
	tip, err := pro.graph.Tip()
	if err != nil {
		return rich.Errorf("could not get tip: %w", err)
	}

	// check the tip is at zero
	if tip.Height != 0 {
		return rich.Errorf("invalid tip height").Uint64("tip_height", tip.Height)
	}

	// vote on the vertex
	err = pro.castVote(tip)
	if err != nil {
		return rich.Errorf("could not cast vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnProposal(proposal *message.Proposal) error {

	// NOTE: the network layer should de-duplicate proposals if we want to
	// avoid expensive double processing of the same proposal multiple times

	// 1) try to confirm the parent vertex of the proposal
	err := pro.confirmParent(proposal)
	if err != nil {
		return rich.Errorf("could not confirm parent: %w", err)
	}

	// 2) try to apply the candidate vertex of the proposal
	err = pro.applyCandidate(proposal)
	if err != nil {
		return rich.Errorf("could not apply candidate: %w", err)
	}

	// 3) extract the proposer vote from the proposal (if we are collector)
	err = pro.extractVote(proposal)
	if err != nil {
		return rich.Errorf("could not extract vote: %w", err)
	}

	// 4) loop back own vote (if we are collector)
	err = pro.loopVote(proposal.Candidate)
	if err != nil {
		return rich.Errorf("could not loop vote: %w", err)
	}

	// 5) cast our own vote for the proposal (if we are not collector)
	err = pro.castVote(proposal.Candidate)
	if err != nil {
		return rich.Errorf("could not cast vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnVote(vote *message.Vote) error {

	// NOTE: the network layer should de-duplicate votes if we want to avoid
	// processing the same vote expensively multiple times

	// 1) collect the vote in our cache
	err := pro.collectVote(vote)
	if err != nil {
		return rich.Errorf("could not collect vote: %w", err)
	}

	// 2) try to build proposal for next round
	err = pro.proposeCandidate(vote.Height, vote.CandidateID)
	if err != nil {
		return rich.Errorf("could not propose candidate: %w", err)
	}

	return nil
}

func (pro *Processor) confirmParent(proposal *message.Proposal) error {

	// 1) validate the quorum signature
	// -> if we don't have a valid quorum for the proposal, we might still want
	// to consider whether the proposal is validly signed and we can punish for
	// an invalid quorum inclusion (we don't do this at the moment)
	err := pro.verify.Quorum(proposal)
	if err != nil {
		return rich.Errorf("could not verify quorum: %w", err)
	}

	// 2) confirm the parent vertex
	// -> as the parent has a qualified majority, we confirm it and don't
	// recheck any of the validity rules; if a non-valid parent can get a quorum
	// our consensus graph state is broken anyway
	err = pro.graph.Confirm(proposal.Candidate.ParentID)
	if err != nil {
		return rich.Errorf("could not confirm parent: %w", err)
	}

	// 3) clear the cache for any pending data up to parent
	// -> the parent height is always equal to the candidate height minus one
	err = pro.cache.Clear(proposal.Candidate.Height - 1)
	if err != nil {
		return rich.Errorf("could not clear cache: %w", err)
	}

	return nil
}

func (pro *Processor) applyCandidate(proposal *message.Proposal) error {

	// 1) check that the proposal vertex is not already in our graph state
	// -> if someone sends us an old proposal for a second time, we can skip
	// processing it here
	stale, err := pro.graph.Contains(proposal.Candidate.ID())
	if err != nil {
		return rich.Errorf("could not check inclusion: %w", err)
	}
	if stale {
		return signal.StaleProposal{Proposal: proposal}
	}

	// 2) check that the proposal is made by the correct leader for the height
	// -> proposals should only ever be made by the valid leader at a given
	// height, so if someone else tries to make one, we should punish them
	leaderID, err := pro.strat.Leader(proposal.Candidate.Height)
	if err != nil {
		return rich.Errorf("could not get leader: %w", err)
	}
	if proposal.Candidate.ProposerID != leaderID {
		return signal.InvalidProposer{Proposal: proposal, Leader: leaderID}
	}

	// 3) check that the proposal is for a height that has not been finalized
	// -> with a safe consensus algorithm, it should be impossible to finalize
	// conflicting proposals, so conflicting proposals are invalid and should be
	// punished where possible
	final, err := pro.graph.Final()
	if err != nil {
		return rich.Errorf("could not get final: %w", err)
	}
	if proposal.Candidate.Height <= final.Height {
		return signal.ConflictingProposal{Proposal: proposal, Final: final}
	}

	// 4) check that the proposal is for a height that is not behind
	// -> this proposal is not necessarily invalid, but it should be impossible
	// to find a majority consensus, as there is already a better candidate that
	// a majority of the network agrees on
	tip, err := pro.graph.Tip()
	if err != nil {
		return rich.Errorf("could not get tip: %w", err)
	}
	if proposal.Candidate.Height < tip.Height {
		return signal.ObsoleteProposal{Proposal: proposal, Tip: tip}
	}

	// 5) check that the proposal has a valid signature
	// -> there is not much we can do if the signature is not valid; however, if
	// there is a valid signature by someone who should not be signing, we can
	// still attribute the mistake and punish
	err = pro.verify.Proposal(proposal)
	if err != nil {
		return rich.Errorf("could not verify proposal: %w", err)
	}

	// 6) try to extend the current graph state with the proposal
	// -> given the previous checks, this should, in theory, always be working;
	// however, we delegate the responsibility for checking what a valid
	// extension is to the external module, which can do additional checks such
	// as validating the payload
	err = pro.graph.Extend(proposal.Candidate)
	if err != nil {
		return rich.Errorf("could not extend graph: %w", err)
	}

	// 7) check if this particular proposal has already been cached, or if
	// there is a double proposal situation being created
	err = pro.cache.Proposal(proposal)
	if err != nil {
		return rich.Errorf("could not cache proposal: %w", err)
	}

	return nil
}

func (pro *Processor) extractVote(proposal *message.Proposal) error {

	// if we are not the collector, no action is required
	selfID, err := pro.sign.Self()
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	collectorID, err := pro.strat.Collector(proposal.Candidate.Height)
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	if selfID != collectorID {
		return nil
	}

	// if we are the collector, process the proposer's vote immediately to give
	// it priority and to make sure that a proposal is generated if the
	// proposer's vote is the only one required to have a qualified majority
	pro.loop.Vote(proposal.Vote())

	return nil
}

func (pro *Processor) loopVote(candidate *base.Vertex) error {

	// if we are the proposer, no action is required, because the vote was
	// already implicitly included in the proposal
	selfID, err := pro.sign.Self()
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	if candidate.ProposerID == selfID {
		return nil
	}

	// if we are not the collector, no action is required, as the vote will be
	// transmitted over the network to the collector
	collectorID, err := pro.strat.Collector(candidate.Height)
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	if collectorID != selfID {
		return nil
	}

	// finally, if we are not the proposer, but the collector, create our own
	// vote for the proposal and immediately process it locally to give it
	// priority and make sure a proposal is generated if our own vote is the
	// only one required for a qualified majority
	vote, err := pro.sign.Vote(candidate)
	if err != nil {
		return rich.Errorf("could not create vote: %w", err)
	}
	pro.loop.Vote(vote)

	return nil
}

func (pro *Processor) castVote(candidate *base.Vertex) error {

	// if we are the proposer, no action is required, as our vote was already
	// implicitly included in the proposal itself
	selfID, err := pro.sign.Self()
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	if candidate.ProposerID == selfID {
		return nil
	}

	// if we are the collector, no action is required, because we already
	// locally processed our own vote and don't need to send it
	collectorID, err := pro.strat.Collector(candidate.Height)
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	if collectorID == selfID {
		return nil
	}

	// otherwise, if we are neither proposer nor collector, we should transmit
	// our vote to the collector over the network
	vote, err := pro.sign.Vote(candidate)
	if err != nil {
		return rich.Errorf("could not create vote: %w", err)
	}
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return rich.Errorf("could not transmit vote: %w", err)
	}

	return nil
}

func (pro *Processor) collectVote(vote *message.Vote) error {

	// 1) discard votes that are on a vertex already included in the state
	contains, err := pro.graph.Contains(vote.CandidateID)
	if err != nil {
		return rich.Errorf("could not check graph inclusion: %w", err)
	}
	if contains {
		return signal.StaleVote{Vote: vote}
	}

	// 2) discard votes on vertices that can't be finalized anymore
	final, err := pro.graph.Final()
	if err != nil {
		return rich.Errorf("could not get final: %w", err)
	}
	if vote.Height <= final.Height {
		return signal.ConflictingVote{Vote: vote, Final: final}
	}

	// 3) ignore votes that are voting on a proposal that is already behind
	// another proposal agreed upon by the network
	tip, err := pro.graph.Tip()
	if err != nil {
		return rich.Errorf("could not get tip: %w", err)
	}
	if vote.Height < tip.Height {
		return signal.ObsoleteVote{Vote: vote, Tip: tip}
	}

	// 4) check if we are the collector for the given vote
	selfID, err := pro.sign.Self()
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	collectorID, err := pro.strat.Collector(vote.Height)
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	if collectorID != selfID {
		return signal.InvalidCollector{Vote: vote, Receiver: selfID, Collector: collectorID}
	}

	// 5) check the signature on the vote
	err = pro.verify.Vote(vote)
	if err != nil {
		return rich.Errorf("could not verify vote signature: %w", err)
	}

	// 6) check if this particular vote has already been processed, or whether
	// it is a double vote situation being created
	err = pro.cache.Vote(vote)
	if err != nil {
		return rich.Errorf("could not cache vote: %w", err)
	}

	return nil
}

func (pro *Processor) proposeCandidate(height uint64, parentID base.Hash) error {

	// 1) check if we have enough votes at the given height and candidate
	threshold, err := pro.strat.Threshold(height)
	if err != nil {
		return rich.Errorf("could not get threshold: %w", err)
	}
	quorum, err := pro.cache.Quorum(height, parentID)
	if err != nil {
		return rich.Errorf("could not build parent: %w", err)
	}
	if uint(len(quorum.SignerIDs)) < threshold {
		return nil
	}

	// 2) create the proposed candidate
	selfID, err := pro.sign.Self()
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	arcID, err := pro.build.Arc()
	if err != nil {
		return rich.Errorf("could not build arc: %w", err)
	}
	candidate := base.Vertex{
		ParentID:   parentID,
		Height:     height + 1,
		ProposerID: selfID,
		ArcID:      arcID,
	}

	// 3) create the proposal and loop it back to ourselves for processing
	proposal, err := pro.sign.Proposal(&candidate)
	if err != nil {
		return rich.Errorf("could not create proposal: %w", err)
	}
	pro.loop.Proposal(proposal)

	// 4) broadcast the proposal to the network
	err = pro.net.Broadcast(proposal)
	if err != nil {
		return rich.Errorf("could not broadcast proposal: %w", err)
	}

	return nil
}
