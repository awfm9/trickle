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
	net    Network  // requires component that implements Network intefrace
	graph  Graph    // requires component that implements Graph intefrace
	build  Builder  // requires component that implements Builder intefrace
	strat  Strategy // requires component that implements Strategy intefrace
	sign   Signer   // requires component that implements Signer intefrace
	verify Verifier // requires component that implements Verifier intefrace
	cache  Cache    // requires component that implements Cache intefrace
	loop   Looper   // requires component that implements Looper intefrace
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
	} //  processor instantiation; fields can be set to anything that implements the corresponding interfaces

	return &pro
}

func (pro *Processor) Bootstrap() error {

	// get current tip of the state
	tip, err := pro.graph.Tip() // obtain the tip of the DAG upon which to build
	// in this function we want to bootstrap from genesis
	if err != nil {
		return rich.Errorf("could not get tip: %w", err)
	}

	// check the tip is at zero
	if tip.Height != 0 {
		// check that the received tip is the first vertex/genesis vertex with zero heigth
		return rich.Errorf("invalid tip height").Uint64("tip_height", tip.Height)
	}

	// vote on the vertex
	err = pro.castVote(tip) // cast own vote
	if err != nil {
		return rich.Errorf("could not cast vote: %w", err)
	}

	return nil
}

// a callback run when a new proposal is recived
func (pro *Processor) OnProposal(proposal *message.Proposal) error {

	// NOTE: the network layer should de-duplicate proposals if we want to
	// avoid expensive double processing of the same proposal multiple times

	// 1) try to confirm the parent vertex of the proposal
	err := pro.confirmParent(proposal)
	if err != nil {
		// this error contains information about the reason for non-confirmation and is eventually handled in the caller scope
		return rich.Errorf("could not confirm parent: %w", err)
	}

	// parent is confirmed -> move to applying the candidate message

	// 2) try to apply the candidate vertex of the proposal
	err = pro.applyCandidate(proposal)
	if err != nil {
		// this error contains information about the reason for not applying candidate and is eventually handled in the caller scope
		return rich.Errorf("could not apply candidate: %w", err)
	}

	// 3) extract the proposer vote from the proposal (if we are collector)
	err = pro.extractVote(proposal)
	// candidate is successfully applied at this point -> so extract vote
	if err != nil {
		// this error contains information about the reason for not extracting a vote and is eventually handled in the caller scope
		return rich.Errorf("could not extract vote: %w", err)
	}

	// 4) loop back own vote (if we are collector)
	err = pro.loopVote(proposal.Candidate)
	if err != nil {
		// this error contains information about the reason for not looping a vote and is eventually handled in the caller scope
		return rich.Errorf("could not loop vote: %w", err)
	}

	// 5) cast our own vote for the proposal (if we are not collector)
	err = pro.castVote(proposal.Candidate)
	if err != nil {
		// this error contains information about the reason for not casting a vote and is eventually handled in the caller scope
		return rich.Errorf("could not cast vote: %w", err)
	}

	return nil
}

// callback run when a vote is received
func (pro *Processor) OnVote(vote *message.Vote) error {

	// NOTE: the network layer should de-duplicate votes if we want to avoid
	// processing the same vote expensively multiple times

	// 1) collect the vote in our cache
	err := pro.collectVote(vote)
	if err != nil {
		// this error contains information about the reason for not collecting a vote and is eventually handled in the caller scope
		return rich.Errorf("could not collect vote: %w", err)
	}

	// 2) try to build proposal for next round
	// clear from comments - create proposal on the received vote
	err = pro.proposeCandidate(vote.Height, vote.CandidateID)
	if err != nil {
		// this error contains information about the reason for not proposing a candidate and is eventually handled in the caller scope
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
	// verify Quorum Certificate of the proposal via Verifier interface
	//
	if err != nil {
		return rich.Errorf("could not verify quorum: %w", err)
	}

	// 2) confirm the parent vertex
	// -> as the parent has a qualified majority, we confirm it and don't
	// recheck any of the validity rules; if a non-valid parent can get a quorum
	// our consensus graph state is broken anyway
	err = pro.graph.Confirm(proposal.Candidate.ParentID)
	// since the proposal is confirmed/accepted by quorum in the previous step, move to confirming parent vertex
	// confirmation steps are abstracted, but generally will use the parent ID to confirm ancestor
	if err != nil {
		return rich.Errorf("could not confirm parent: %w", err)
	}

	// 3) clear the cache for any pending data up to parent
	// -> the parent height is always equal to the candidate height minus one
	err = pro.cache.Clear(proposal.Candidate.Height - 1)
	// clear cache excluding parent height; this is clear from comments also
	if err != nil {
		return rich.Errorf("could not clear cache: %w", err)
	}

	// up to this step - proposal has a quorum certificate verified and confirmed parent,
	// cache is successfully cleared out before parent -> return to caller scope
	return nil
}

func (pro *Processor) applyCandidate(proposal *message.Proposal) error {

	// 1) check that the proposal vertex is not already in our graph state
	// -> if someone sends us an old proposal for a second time, we can skip
	// processing it here
	stale, err := pro.graph.Contains(proposal.Candidate.ID())
	// check if the processor instance has the proposal candidate ID in its graph history
	// if it is included -> stop procedure and transmit a signal that the proposal is stale
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
	// get leader ID for the candidate vertex height
	if err != nil {
		return rich.Errorf("could not get leader: %w", err)
	}
	// if the leader ID included in the candidate height is not the same as the candidate proposer ID
	// -> return a signal for invalid proposer received
	// because the leader is the one that sends out the proposals to nodes (star communication)
	if proposal.Candidate.ProposerID != leaderID {
		return signal.InvalidProposer{Proposal: proposal, Leader: leaderID}
	}

	// 3) check that the proposal is for a height that has not been finalized
	// -> with a safe consensus algorithm, it should be impossible to finalize
	// conflicting proposals, so conflicting proposals are invalid and should be
	// punished where possible
	final, err := pro.graph.Final()
	// get last finalized vertex/transaction of the graph
	if err != nil {
		return rich.Errorf("could not get final: %w", err)
	}
	// compare candidate vertex height from the received proposal with the last finalized vertex height
	// and if the candidate height is not higher than the finalized vertex height
	// -> send a signal for conflicting proposal (double spending is one of the conflict types)
	// this istance knows about a higher view in the graph than the proposer sender
	if proposal.Candidate.Height <= final.Height {
		return signal.ConflictingProposal{Proposal: proposal, Final: final}
	}

	// 4) check that the proposal is for a height that is not behind
	// -> this proposal is not necessarily invalid, but it should be impossible
	// to find a majority consensus, as there is already a better candidate that
	// a majority of the network agrees on
	tip, err := pro.graph.Tip()
	// get curent tip of correct path via Graph interface implementation
	if err != nil {
		return rich.Errorf("could not get tip: %w", err)
	}
	// compare the current tip height with the proposal candidate height ->
	// theoretically quorum signature should have been received for the best candidate
	// in the past height => if the candidate is for a height before the current tip ->
	// send obsolete proposal signal
	if proposal.Candidate.Height < tip.Height {
		return signal.ObsoleteProposal{Proposal: proposal, Tip: tip}
	}

	// 5) check that the proposal has a valid signature
	// -> there is not much we can do if the signature is not valid; however, if
	// there is a valid signature by someone who should not be signing, we can
	// still attribute the mistake and punish
	err = pro.verify.Proposal(proposal)
	// this is logically clear and also comment explains it
	// verify signature using Verifier interface
	if err != nil {
		return rich.Errorf("could not verify proposal: %w", err)
	}

	// 6) try to extend the current graph state with the proposal
	// -> given the previous checks, this should, in theory, always be working;
	// however, we delegate the responsibility for checking what a valid
	// extension is to the external module, which can do additional checks such
	// as validating the payload
	err = pro.graph.Extend(proposal.Candidate)
	// previous checks are successful to this point -> rest - including payload validation
	// are handled by the external component implementing Graph interface
	// explained also by the comment
	if err != nil {
		return rich.Errorf("could not extend graph: %w", err)
	}

	// 7) check if this particular proposal has already been cached, or if
	// there is a double proposal situation being created
	err = pro.cache.Proposal(proposal)
	// not sure what does the Proposal method for Cache interface should do in this case
	// but it should check if there is a double record in the cache record and add it if not present
	if err != nil {
		return rich.Errorf("could not cache proposal: %w", err)
	}

	return nil
}

func (pro *Processor) extractVote(proposal *message.Proposal) error {

	// if we are not the collector, no action is required
	selfID, err := pro.sign.Self()
	// get processor instance self id as [32]byte via Signer interface implementation
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	collectorID, err := pro.strat.Collector(proposal.Candidate.Height)
	// get collector ID from received proposal candidate height via Strategy interface implementaion
	// -> collector of new view messages from this height
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	if selfID != collectorID {
		// return to caller scope if current processor instance is not the collector
		return nil
	}

	// if we are the collector, process the proposer's vote immediately to give
	// it priority and to make sure that a proposal is generated if the
	// proposer's vote is the only one required to have a qualified majority
	pro.loop.Vote(proposal.Vote())
	// if the current instance is a collector ->
	// process the proposal vote via Looper interface implementation
	// (collect new view mesages)

	return nil
}

func (pro *Processor) loopVote(candidate *base.Vertex) error {

	// if we are the proposer, no action is required, because the vote was
	// already implicitly included in the proposal
	selfID, err := pro.sign.Self()
	// get processor instance self id as [32]byte via Signer interface implementation

	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	// compare candidate proposer ID with processor instance self ID
	// -> if current processor is the proposer -> return to caller scope
	// because it had included its vote in the proposal it created and has been extracted if this node is collector
	// (this is explained in the comments also)

	// the leader receives new messages from replicas; each of the messages contain a Quorum Certificate
	// for the replica's highest view on the graph
	// the leader selects the message with the QC for the highest view among them
	// and extends its tip with a new proposal that it broadcasts
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

	// up to this step - the current instance is a collector
	// the node signs the candidate vertex with its key and sends out its vote for it
	// using the Signer interface implementation
	// after the vote, the node processes its own vote immediately

	// finally, if we are not the proposer, but the collector, create our own
	// vote for the proposal and immediately process it locally to give it
	// priority and make sure a proposal is generated if our own vote is the
	// only one required for a qualified majority - I do not understand this part
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
	selfID, err := pro.sign.Self() // get own ID as a [32]byte
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	// compare own ID with the received canidate Proposer ID -> if the same
	// return to previous scope because this processor receiver instance is the proposer and will not vote
	//
	if candidate.ProposerID == selfID {
		return nil
	}

	// if we are the collector, no action is required, because we already
	// locally processed our own vote and don't need to send it
	collectorID, err := pro.strat.Collector(candidate.Height) // get collector saved ID
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	// compare own ID with the collectior saved ID in the processor receiver
	// if own ID is the collector ID -> the processor instance is the collector in the
	// strategy (protocol) and returns to caller scope
	// current instance processed its vote in the previous step in `loopVote`
	if collectorID == selfID {
		return nil
	}

	// after verified this processor is not collector or proposer ->
	// proceed with voting on the received candidate

	// otherwise, if we are neither proposer nor collector, we should transmit
	// our vote to the collector over the network
	vote, err := pro.sign.Vote(candidate)
	// sign candidate with own signature (nodes are represented by their keys)
	if err != nil {
		return rich.Errorf("could not create vote: %w", err)
	}
	// after signing is done with no errors -> transmit vote to network (explained also in the original comments)
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return rich.Errorf("could not transmit vote: %w", err)
	}

	// all steps are performed successfully -> return nil
	return nil
}

func (pro *Processor) collectVote(vote *message.Vote) error {

	// 1) discard votes that are on a vertex already included in the state
	contains, err := pro.graph.Contains(vote.CandidateID)
	// check if the processor instance has the proposal candidate ID in its graph history
	// if it is included -> stop procedure and transmit a signal that the proposal is stale
	if err != nil {
		return rich.Errorf("could not check graph inclusion: %w", err)
	}
	if contains {
		return signal.StaleVote{Vote: vote}
	}

	// 2) discard votes on vertices that can't be finalized anymore
	final, err := pro.graph.Final()
	// get finalized vertex/transaction of the graph
	if err != nil {
		return rich.Errorf("could not get final: %w", err)
	}
	// compare final vertex height with vote height ->
	// if the vote is for a height that is not higher than the finalized one - there is a conflict
	// between the vote and some of the existing vertices and history cannot be rewritten
	// -> send signal to the network and return to caller scope
	if vote.Height <= final.Height {
		return signal.ConflictingVote{Vote: vote, Final: final}
	}

	// 3) ignore votes that are voting on a proposal that is already behind
	// another proposal agreed upon by the network
	tip, err := pro.graph.Tip()
	// get current tip of the graph correct path via the Graph interface implementation
	if err != nil {
		return rich.Errorf("could not get tip: %w", err)
	}
	// if the vote received is for a height that is behind the current tip ->
	// it is doubling some vertext that has theoretically received its best voting result
	// so send a signal for obsolete vote to the network, which implicitly means in the case
	// that this instance has knowledge of higer view than the one in the vote or the vote is fraudulent
	if vote.Height < tip.Height {
		return signal.ObsoleteVote{Vote: vote, Tip: tip}
	}

	// 4) check if we are the collector for the given vote
	selfID, err := pro.sign.Self()
	// get processor instance self id as [32]byte via Signer interface implementation
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	collectorID, err := pro.strat.Collector(vote.Height)
	// get collector ID via Strategy interface
	if err != nil {
		return rich.Errorf("could not get collector: %w", err)
	}
	// if this processor instance is not the collector, return to caller scope and send a signal that
	// it is not the collector if this vote
	if collectorID != selfID {
		return signal.InvalidCollector{Vote: vote, Receiver: selfID, Collector: collectorID}
	}

	// 5) check the signature on the vote
	// since this instance is the collector of the vote -> proceed -> verify it via the Verifier interface implementation
	err = pro.verify.Vote(vote)
	if err != nil {
		return rich.Errorf("could not verify vote signature: %w", err)
	}

	// 6) check if this particular vote has already been processed, or whether
	// it is a double vote situation being created
	err = pro.cache.Vote(vote)
	// check if the vote exists in the cache records via Cache interface implementation
	// not sure how is the cache set up and up to how many votes it can check, but I assume
	// the implementation will look for this vote and add it if it is not present in the slice
	if err != nil {
		return rich.Errorf("could not cache vote: %w", err)
	}

	return nil
}

func (pro *Processor) proposeCandidate(height uint64, parentID base.Hash) error {

	// 1) check if we have enough votes at the given height and candidate
	// executed by the Strategy interface
	// we need to check if the threshold of partial signatures on a message by the replicas
	// sent to the leader
	// if the threshold is reached - we do not neet all replicas to sign
	// in order to generate the complete signature
	threshold, err := pro.strat.Threshold(height)
	if err != nil {
		return rich.Errorf("could not get threshold: %w", err)
	}
	quorum, err := pro.cache.Quorum(height, parentID)
	// obtained via Cache interface implementation
	// calculate Quorum Certificate for the height received
	if err != nil {
		return rich.Errorf("could not build parent: %w", err)
	}
	// if the required threshold of replicas partial signatures is not reached
	// return to caller scope without doing anything because
	// protocol requires it
	if uint(len(quorum.SignerIDs)) < threshold {
		return nil
	}

	// 2) create the proposed candidate
	selfID, err := pro.sign.Self()
	// get own ID as [32]byte
	if err != nil {
		return rich.Errorf("could not get self: %w", err)
	}
	arcID, err := pro.build.Arc()
	// build arcID between child and parent vertices via Builder interface implementation
	if err != nil {
		return rich.Errorf("could not build arc: %w", err)
	}
	// create candidate for the next height
	// including the parent and proposer IDs
	candidate := base.Vertex{
		ParentID:   parentID,
		Height:     height + 1,
		ProposerID: selfID,
		ArcID:      arcID,
	}

	// 3) create the proposal and loop it back to ourselves for processing
	proposal, err := pro.sign.Proposal(&candidate)
	// sign proposal with own key
	if err != nil {
		return rich.Errorf("could not create proposal: %w", err)
	}
	// loop proposal in order to process it without waiting
	pro.loop.Proposal(proposal)

	// 4) broadcast the proposal to the network
	// clear from the original comments
	err = pro.net.Broadcast(proposal)
	if err != nil {
		return rich.Errorf("could not broadcast proposal: %w", err)
	}

	return nil
}
