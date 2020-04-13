package consensus

import (
	"fmt"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Processor struct {
	net    Network
	graph  Graph
	build  Builder
	strat  Strategy
	sign   Signer
	verify Verifier
	pcache ProposalCache
	vcache VoteCache
	Round  uint64
}

func NewProcessor(net Network, graph Graph, build Builder, strat Strategy, sign Signer, verify Verifier, pcache ProposalCache, vcache VoteCache) *Processor {

	pro := Processor{
		net:    net,
		graph:  graph,
		build:  build,
		strat:  strat,
		sign:   sign,
		verify: verify,
		pcache: pcache,
		vcache: vcache,
		Round:  0,
	}

	return &pro
}

func (pro *Processor) Bootstrap(arcID model.Hash) error {

	// check if we are still at zero round
	if pro.Round != 0 {
		return fmt.Errorf("graph round not zero")
	}

	// create root vertex
	root := model.Vertex{
		Parent:   nil,
		Height:   0,
		ArcID:    arcID,
		SignerID: model.ZeroHash,
	}

	// extend the graph with the root
	err := pro.graph.Extend(&root)
	if err != nil {
		return fmt.Errorf("could not extend graph with root: %w", err)
	}

	// create the vote for the proposed vertex
	vote, err := pro.sign.Vote(&root)
	if err != nil {
		return fmt.Errorf("could not create genesis vote: %w", err)
	}

	// get the collector of the genesis round
	collectorID, err := pro.strat.Collector(root.Height)
	if err != nil {
		return fmt.Errorf("could not get collector: %w", err)
	}

	// send the vote to the genesis collector
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit genesis vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnProposal(proposal *message.Proposal) error {

	// check if the proposal falls within the finalized state
	final, exists := pro.graph.Final()
	if exists && proposal.Height <= final.Height {
		return ConflictingProposal{Proposal: proposal, Final: final.Height}
	}

	// get the proposer at the proposal height
	leaderID, err := pro.strat.Leader(proposal.Height)
	if err != nil {
		return fmt.Errorf("could not get proposer: %w", err)
	}

	// check if the proposal is signed by correct proposer
	if proposal.SignerID != leaderID {
		return InvalidProposer{Proposal: proposal, Leader: leaderID}
	}

	// check if the proposed vertex has valid signature and parent
	err = pro.verify.Proposal(proposal)
	if err != nil {
		return fmt.Errorf("could not verify proposal: %w", err)
	}

	// check if we already had a proposal by this proposer at this height
	fresh, err := pro.pcache.Store(proposal)
	if err != nil {
		return fmt.Errorf("could not buffer proposal: %w", err)
	}
	if !fresh {
		return nil
	}

	// set the current round if the confirmed height increased
	if proposal.Height > pro.Round {
		pro.Round = proposal.Height
	}

	// NOTE: We currently don't check if we have the parent, which allows us to
	// skip ahead, even if we don't know the chain up to this point.

	// NOTE: We never check if the parent is on a vertex that is a valid
	// extension of the state; if it wasn't, the system would already be
	// compromised. This allows us to skip ahead to the next height regardless
	// of the validity of the new proposal.

	// confirm the parent of the proposal
	err = pro.graph.Confirm(proposal.Parent.VertexID)
	if err != nil {
		return fmt.Errorf("could not confirm parent: %w", err)
	}

	// clear the cache for votes up to the confirmed height
	err = pro.vcache.Clear(proposal.Parent.Height)
	if err != nil {
		return fmt.Errorf("could not clear vote cache: %w", err)
	}

	// clear the cache for proposals up to the confirmed height
	err = pro.pcache.Clear(proposal.Parent.Height)
	if err != nil {
		return fmt.Errorf("could not clear proposal cache: %w", err)
	}

	// check if the new proposal is a valid extension of the graph
	err = pro.graph.Extend(proposal.Vertex)
	if err != nil {
		return fmt.Errorf("could not extend graph: %w", err)
	}

	// check that the proposal is not already outdated; we don't want to vote
	// on proposals that are already worse than a pending proposal with majority
	cutoff := pro.Round
	if proposal.Height < cutoff {
		return ObsoleteProposal{Proposal: proposal, Cutoff: cutoff}
	}

	// get own ID to compare against collector and leader
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}

	// get the collector for the proposal round
	collectorID, err := pro.strat.Collector(proposal.Height)
	if err != nil {
		return fmt.Errorf("could not get collector: %w", err)
	}

	// if we are the next collector, we should collect the vote that is included
	// implicitly in the vertex proposal by the proposal; in order to avoid
	// creating extra code paths, we send it as message to ourselves
	if collectorID == selfID {
		err = pro.net.Transmit(proposal.Vote(), selfID)
		if err != nil {
			return fmt.Errorf("could not transmit proposer vote to self: %w", err)
		}
	}

	// if we are the current proposer, the vote was already implicitly included
	// in the proposal, so we don't need to transmit it again
	if proposal.SignerID == selfID {
		return nil
	}

	// create own vote for the proposed vertex
	vote, err := pro.sign.Vote(proposal.Vertex)
	if err != nil {
		return fmt.Errorf("could not create vote: %w", err)
	}

	// send the vote for the proposed vertex to the next collector
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnVote(vote *message.Vote) error {

	// check if the vote falls within the finalized state
	final, exists := pro.graph.Final()
	if exists && vote.Height <= final.Height {
		return ConflictingVote{Vote: vote, Final: final.Height}
	}

	// check if the vote is already outdated; we don't want to build proposals
	// that, even if they get a majority, are already behind another path
	cutoff := pro.Round - 1
	if cutoff > pro.Round {
		cutoff = pro.Round
	}
	if vote.Height < cutoff {
		return ObsoleteVote{Vote: vote, Cutoff: cutoff}
	}

	// get own ID to compare against collector
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}

	// get the collector for the vote
	collectorID, err := pro.strat.Collector(vote.Height)
	if err != nil {
		return fmt.Errorf("could not get vote collector: %w", err)
	}

	// check if we are the collector for the round
	if collectorID != selfID {
		return InvalidCollector{Vote: vote, Collector: collectorID, Receiver: selfID}
	}

	// check if the vote has a valid signature
	err = pro.verify.Vote(vote)
	if err != nil {
		return fmt.Errorf("could not verify vote signature: %w", err)
	}

	// check if we already have a vote by this voter
	fresh, err := pro.vcache.Store(vote)
	if err != nil {
		return fmt.Errorf("could not tally vote: %w)", err)
	}
	if !fresh {
		return nil
	}

	// get the threshold (for all rounds currently)
	threshold, err := pro.strat.Threshold(vote.Height)
	if err != nil {
		return fmt.Errorf("could not get threshold: %w", err)
	}

	// get the votes for the given vertex
	votes, err := pro.vcache.Retrieve(vote.Height, vote.VertexID)
	if err != nil {
		return fmt.Errorf("could not get votes: %w", err)
	}

	// check if we have reached the threshold
	if uint(len(votes)) <= threshold {
		return nil
	}

	// collect vote signers and signatures
	var signature []byte
	signerIDs := make([]model.Hash, 0, len(votes))
	for _, vote := range votes {
		signerIDs = append(signerIDs, vote.SignerID)
		signature = append(signature, vote.Signature...)
	}

	// get a random arc from the builder
	arcID, err := pro.build.Arc()
	if err != nil {
		return fmt.Errorf("could not build arc: %w", err)
	}

	// NOTE: this can create a parent with a vertex have not even seen yet;
	// however, with the majority voting for it, we should be able to rely on it

	// create the Parent for the new proposal
	parent := model.Parent{
		VertexID:  vote.VertexID,
		SignerIDs: signerIDs,
		Signature: signature,
	}

	// create the vertex for the new proposal
	vertex := model.Vertex{
		Height: vote.Height + 1,
		Parent: &parent,
		ArcID:  arcID,
	}

	// create the new proposal
	proposal, err := pro.sign.Proposal(&vertex)
	if err != nil {
		return fmt.Errorf("could not create proposal: %w", err)
	}

	// we trust our proposal, so we set the new round redundantly here
	pro.Round = proposal.Height

	// broadcast the proposal to the network
	// NOTE: the network module should short-circuit one copy of this messae to
	// ourselves, which will lead to the state transition to the next height
	err = pro.net.Broadcast(proposal)
	if err != nil {
		return fmt.Errorf("could not broadcast proposal: %w", err)
	}

	return nil
}
