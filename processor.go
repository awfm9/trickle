package consensus

import (
	"fmt"

	"github.com/alvalor/consensus/errors"
	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Processor struct {
	net       Network
	graph     Graph
	build     Builder
	strat     Strategy
	sign      Signer
	verify    Verifier
	proposals ProposalCache
	votes     VoteCache
}

func NewProcessor(net Network, graph Graph, build Builder, strat Strategy, sign Signer, verify Verifier, proposals ProposalCache, votes VoteCache) *Processor {

	pro := Processor{
		net:       net,
		graph:     graph,
		build:     build,
		strat:     strat,
		sign:      sign,
		verify:    verify,
		proposals: proposals,
		votes:     votes,
	}

	return &pro
}

func (pro *Processor) Bootstrap() error {

	// get current tip of the state
	tip, err := pro.graph.Tip()
	if err != nil {
		return fmt.Errorf("could not get tip: %w", err)
	}

	// check the tip is at zero
	if tip.Height != 0 {
		return fmt.Errorf("invalid bootstrap height (%d)", tip.Height)
	}

	// vote on the vertex
	err = pro.voteVertex(tip)
	if err != nil {
		return fmt.Errorf("could not vote on genesis: %w", err)
	}

	return nil
}

func (pro *Processor) OnProposal(proposal *message.Proposal) error {

	// 1) check if proposal is already pending
	fresh, err := pro.proposals.Store(proposal)
	if err != nil {
		return fmt.Errorf("could not buffer proposal: %w", err)
	}
	if !fresh {
		return nil
	}

	// 2) check the proposal is consistent within itself
	err = pro.validateProposal(proposal)
	if err != nil {
		return fmt.Errorf("could not validate proposal: %w", err)
	}

	// 3) confirm the referenced parent vertex
	err = pro.confirmVertex(proposal.Parent)
	if err != nil {
		return fmt.Errorf("could not confirm parent: %w", err)
	}

	// 4) check the proposal is valid extension of state
	err = pro.applyVertex(proposal.Vertex)
	if err != nil {
		return fmt.Errorf("could not apply proposal: %w", err)
	}

	// 5) vote on new proposed vertex
	err = pro.voteVertex(proposal.Vertex)
	if err != nil {
		return fmt.Errorf("could not vote proposal: %w", err)
	}

	return nil
}

func (pro *Processor) OnVote(vote *message.Vote) error {

	// 1) check if vote is already pending
	fresh, err := pro.votes.Store(vote)
	if err != nil {
		return fmt.Errorf("could not buffer vote: %w", err)
	}
	if !fresh {
		return nil
	}

	// 2) validate the vote internally
	err = pro.validateVote(vote)
	if err != nil {
		return fmt.Errorf("could not validate vote: %w", err)
	}

	// 3) try to build proposal
	err = pro.proposeVertex(vote.Height)
	if err != nil {
		return fmt.Errorf("could not build proposal: %w", err)
	}

	return nil
}

func (pro *Processor) validateProposal(proposal *message.Proposal) error {

	// 1) check that the proposal was not made by ourselves
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}
	if proposal.SignerID == selfID {
		return fmt.Errorf("validating proposal by self")
	}

	// 2) check that the proposal is made by the correct leader for the height
	leaderID, err := pro.strat.Leader(proposal.Height)
	if err != nil {
		return fmt.Errorf("could not get proposer: %w", err)
	}
	if proposal.SignerID != leaderID {
		return errors.InvalidProposer{Proposal: proposal, Leader: leaderID}
	}

	// 3) check that the proposal doesn't conflict with immutable graph state
	final, err := pro.graph.Final()
	if err != nil {
		return fmt.Errorf("could not get final: %w", err)
	}
	if proposal.Height <= final.Height {
		return errors.ConflictingProposal{Proposal: proposal, Final: final.Height}
	}

	// 4) check that the parent and proposal signatures are correct
	err = pro.verify.Proposal(proposal)
	if err != nil {
		return fmt.Errorf("could not verify proposal: %w", err)
	}

	return nil
}

func (pro *Processor) confirmVertex(ref *model.Reference) error {

	// 1) confirm the vertex in the graph
	err := pro.graph.Confirm(ref.VertexID)
	if err != nil {
		return fmt.Errorf("could not confirm parent: %w", err)
	}

	// 2) clear volatile data no longer needed
	err = pro.votes.Clear(ref.Height)
	if err != nil {
		return fmt.Errorf("could not clear vote cache: %w", err)
	}
	err = pro.proposals.Clear(ref.Height)
	if err != nil {
		return fmt.Errorf("could not clear proposal cache: %w", err)
	}

	return nil
}

func (pro *Processor) applyVertex(vertex *model.Vertex) error {

	// NOTE: We never check whether the parent was a valid extension of the
	// state; it already has a qualified majority, so if it's not valid, it
	// means the graph state is already compromised. This assumption allows us
	// to skip ahead to a proposal height regardless of how many vertices we
	// are missing in-between.

	// 1) check if the vertex is already obsolete; we should probably not waste
	// time processing a vertex that is conflicting with another vertex that
	// already had the confirmation of the majority; it means that no one will
	// choose this vertex to build on anyway
	tip, err := pro.graph.Tip()
	if err != nil {
		return fmt.Errorf("could not get tip: %w", err)
	}
	if vertex.Height < tip.Height {
		return errors.ObsoleteProposal{Vertex: vertex, Cutoff: tip.Height}
	}

	// 2) check if the vertex is a valid extension of the graph state and
	// introduce it into the state if it is
	err = pro.graph.Extend(vertex)
	if err != nil {
		return fmt.Errorf("could not extend graph: %w", err)
	}

	return nil
}

func (pro *Processor) voteVertex(vertex *model.Vertex) error {

	// 1) check if we are the proposer and we can skip voting
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}
	if vertex.SignerID == selfID {
		return nil
	}

	// 2) create our vote for the proposed vertex
	vote, err := pro.sign.Vote(vertex)
	if err != nil {
		return fmt.Errorf("could not create vote: %w", err)
	}

	// 3) send the vote to the collector for the proposed vertex
	collectorID, err := pro.strat.Collector(vertex.Height)
	if err != nil {
		return fmt.Errorf("could not get collector: %w", err)
	}
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit vote: %w", err)
	}

	return nil
}

func (pro *Processor) validateVote(vote *message.Vote) error {

	// 1) check that vote is not our own
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}
	if vote.SignerID == selfID {
		return fmt.Errorf("validating vote by self")
	}

	// 1) discard votes on vertices already introduced to the graph state
	contains, err := pro.graph.Contains(vote.VertexID)
	if err != nil {
		return fmt.Errorf("could not check graph: %w", err)
	}
	if contains {
		return errors.StaleVote{Vote: vote}
	}

	// 2) discard votes on vertices conflicting with immutable graph state
	final, err := pro.graph.Final()
	if err != nil {
		return fmt.Errorf("could not get final: %w", err)
	}
	if vote.Height <= final.Height {
		return errors.ConflictingVote{Vote: vote, Final: final.Height}
	}

	// 3) discard votes building a proposal that is already behind; we don't
	// want to waste time building a proposal that is below the height of a
	// proposal everyone has already seen and confirmed once
	tip, err := pro.graph.Tip()
	if err != nil {
		return fmt.Errorf("could not get tip: %w", err)
	}
	if vote.Height < tip.Height {
		return errors.ObsoleteVote{Vote: vote, Cutoff: tip.Height}
	}

	// 4) check if we are the collector for the given vote
	collectorID, err := pro.strat.Collector(vote.Height)
	if err != nil {
		return fmt.Errorf("could not get vote collector: %w", err)
	}
	if collectorID != selfID {
		return errors.InvalidCollector{Vote: vote, Collector: collectorID, Receiver: selfID}
	}

	// 5) check the signature on the vote
	err = pro.verify.Vote(vote)
	if err != nil {
		return fmt.Errorf("could not verify vote signature: %w", err)
	}

	return nil
}

func (pro *Processor) proposeVertex(height uint64) error {

	// 1) check if we have enough votes at the given height
	threshold, err := pro.strat.Threshold(height)
	if err != nil {
		return fmt.Errorf("could not get threshold: %w", err)
	}
	vertexID, votes, err := pro.votes.Retrieve(height)
	if err != nil {
		return fmt.Errorf("could not get votes: %w", err)
	}
	if uint(len(votes)) <= threshold {
		return nil
	}

	// NOTE: this can create a parent with a vertex we have not even seen yet;
	// however, with the majority voting for it, we should be able to rely on it

	// 2) create and confirm parent reference for the proposed vertex
	var signature []byte
	signerIDs := make([]model.Hash, 0, len(votes))
	for _, vote := range votes {
		signerIDs = append(signerIDs, vote.SignerID)
		signature = append(signature, vote.Signature...)
	}
	parent := model.Reference{
		Height:    height - 1,
		VertexID:  vertexID,
		SignerIDs: signerIDs,
		Signature: signature,
	}
	err = pro.confirmVertex(&parent)
	if err != nil {
		return fmt.Errorf("could not confirm parent: %w", err)
	}

	// 3) create and process the proposed vertex
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}
	arcID, err := pro.build.Arc()
	if err != nil {
		return fmt.Errorf("could not build arc: %w", err)
	}
	vertex := model.Vertex{
		Parent:   &parent,
		Height:   height,
		ArcID:    arcID,
		SignerID: selfID,
	}
	err = pro.applyVertex(&vertex)
	if err != nil {
		return fmt.Errorf("could not apply vertex: %w", err)
	}

	// 4) sign and broadcast the proposed vertex
	proposal, err := pro.sign.Proposal(&vertex)
	if err != nil {
		return fmt.Errorf("could not create proposal: %w", err)
	}
	err = pro.net.Broadcast(proposal)
	if err != nil {
		return fmt.Errorf("could not broadcast proposal: %w", err)
	}

	return nil
}
