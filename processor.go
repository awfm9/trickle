package consensus

import (
	"fmt"
	"time"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Processor struct {
	net    Network
	state  State
	build  Builder
	sign   Signer
	verify Verifier
	buffer Buffer
}

func NewProcessor(state State, net Network, build Builder, sign Signer, verify Verifier, buffer Buffer) *Processor {

	pro := Processor{
		state:  state,
		net:    net,
		build:  build,
		sign:   sign,
		verify: verify,
		buffer: buffer,
	}

	return &pro
}

func (pro *Processor) Bootstrap(genesis *model.Block) error {

	// check that we are at height zero
	round, err := pro.state.Round()
	if err != nil {
		return fmt.Errorf("could not get round: %w", err)
	}

	// check if round is zero
	if round != 0 {
		return fmt.Errorf("invalid round for bootstrap (%d)", round)
	}

	// check that genesis block is at height zero
	if genesis.Height != 0 {
		return fmt.Errorf("invalid genesis height (%d)", genesis.Height)
	}

	// check that genesis has no QC
	if genesis.QC != nil {
		return fmt.Errorf("genesis has parent (%x)", genesis.QC.BlockID)
	}

	// check that genesis has no payload
	if genesis.PayloadHash != model.ZeroHash {
		return fmt.Errorf("genesis has payload (%x)", genesis.PayloadHash)
	}

	// check that genesis block has no proposer
	if genesis.SignerID != model.ZeroHash {
		return fmt.Errorf("genesis has signer (%x)", genesis.SignerID)
	}

	// create the vote for the proposed block
	vote, err := pro.sign.Vote(genesis)
	if err != nil {
		return fmt.Errorf("could not create genesis vote: %w", err)
	}

	// get the collector of the genesis round
	collectorID, err := pro.state.Leader(genesis.Height + 1)
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

	// get the current round
	cutoff, err := pro.state.Round()
	if err != nil {
		return fmt.Errorf("could not get cutoff: %w", err)
	}

	// check if the proposal is not outdated
	if proposal.Height < cutoff {
		return ObsoleteProposal{Proposal: proposal, Cutoff: cutoff}
	}

	// get the proposer at the proposal height
	leaderID, err := pro.state.Leader(proposal.Height)
	if err != nil {
		return fmt.Errorf("could not get proposer: %w", err)
	}

	// check if the proposal is signed by correct proposer
	if proposal.SignerID != leaderID {
		return InvalidProposer{Proposal: proposal, Leader: leaderID}
	}

	// check if the proposed block has a valid signature & QC
	err = pro.verify.Proposal(proposal)
	if err != nil {
		return fmt.Errorf("could not verify proposal: %w", err)
	}

	// check if we already had a proposal by this proposer
	fresh, err := pro.buffer.Proposal(proposal)
	if err != nil {
		return fmt.Errorf("could not buffer proposal: %w", err)
	}
	if !fresh {
		return nil
	}

	// NOTE: we currently don't check if we have the parent, which allows us to
	// skip ahead, even if we don't know the chain up to this point

	// NOTE: we never check if the QC is on a block that is a valid extension of
	// the state; if it wasn't, the system would already be compromised, because
	// we have a majority of validators voting for an invalid block - we can
	// therefore simply jump to the height after the QC/parent

	// set our state to the new height
	err = pro.state.Set(proposal.Height)
	if err != nil {
		return fmt.Errorf("could not transition round: %w", err)
	}

	// clear the buffer for the voted block
	err = pro.buffer.Clear(proposal.Height - 1)
	if err != nil {
		return fmt.Errorf("could not clear buffer: %w", err)
	}

	// TODO: check if the proposed block is a valid extension of the state

	// get own ID
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}

	// get the collector for the proposal round
	collectorID, err := pro.state.Leader(proposal.Height + 1)
	if err != nil {
		return fmt.Errorf("could not get collector: %w", err)
	}

	// if we are the next collector, we should collect the vote that is included
	// implicitly in the block proposal by the proposal; in order to avoid
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

	// create own vote for the proposed block
	vote, err := pro.sign.Vote(proposal.Block)
	if err != nil {
		return fmt.Errorf("could not create vote: %w", err)
	}

	// send the vote for the proposed block to the next collector
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnVote(vote *message.Vote) error {

	// get current round
	cutoff, err := pro.state.Round()
	if err != nil {
		return fmt.Errorf("could not get cutoff: %w", err)
	}

	// check that the vote is not outdated
	if vote.Height < cutoff {
		return ObsoleteVote{Vote: vote, Cutoff: cutoff}
	}

	// get own ID
	selfID, err := pro.sign.Self()
	if err != nil {
		return fmt.Errorf("could not get self: %w", err)
	}

	// get the collector for the vote
	collectorID, err := pro.state.Leader(vote.Height + 1)
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
	fresh, err := pro.buffer.Vote(vote)
	if err != nil {
		return fmt.Errorf("could not tally vote: %w)", err)
	}
	if !fresh {
		return nil
	}

	// get the threshold (for all rounds currently)
	threshold, err := pro.state.Threshold()
	if err != nil {
		return fmt.Errorf("could not get threshold: %w", err)
	}

	// get the votes for the given block
	votes, err := pro.buffer.Votes(vote.BlockID)
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

	// get a random payload
	payloadHash, err := pro.build.PayloadHash()
	if err != nil {
		return fmt.Errorf("could not build payload: %w", err)
	}

	// NOTE: this can create a QC for a block have not even seen yet;
	// however, with the majority voting for it, we should be able to rely on it

	// create the QC for the new proposal
	qc := model.QC{
		BlockID:   vote.BlockID,
		SignerIDs: signerIDs,
		Signature: signature,
	}

	// create the block for the new proposal
	block := model.Block{
		Height:      vote.Height + 1,
		QC:          &qc,
		PayloadHash: payloadHash,
		Timestamp:   time.Now().UTC(),
	}

	// create the new proposal
	proposal, err := pro.sign.Proposal(&block)
	if err != nil {
		return fmt.Errorf("could not create proposal: %w", err)
	}

	// broadcast the proposal to the network
	// NOTE: the network module should short-circuit one copy of this messae to
	// ourselves, which will lead to the state transition to the next height
	err = pro.net.Broadcast(proposal)
	if err != nil {
		return fmt.Errorf("could not broadcast proposal: %w", err)
	}

	return nil
}
