package consensus

import (
	"fmt"
	"time"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type Processor struct {
	db     Database
	net    Network
	state  State
	sign   Signer
	verify Verifier
	buf    Buffer
	build  Builder
}

func NewProcessor(db Database, net Network, state State, sign Signer, verify Verifier, buf Buffer, build Builder) *Processor {

	pro := Processor{
		db:     db,
		net:    net,
		state:  state,
		sign:   sign,
		verify: verify,
		buf:    buf,
		build:  build,
	}

	return &pro
}

func (pro *Processor) Bootstrap(genesis *model.Block) error {

	// check that we are at height zero
	round := pro.state.Round()
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

	// store the genesis block
	err := pro.db.Store(genesis)
	if err != nil {
		return fmt.Errorf("could not store genesis: %w", err)
	}

	// create the vote for the proposed block
	vote, err := pro.sign.Vote(genesis)
	if err != nil {
		return fmt.Errorf("could not create genesis vote: %w", err)
	}

	// send the vote to the very first leader
	collectorID := pro.state.Leader(genesis.Height + 1)
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit genesis vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnProposal(proposal *message.Proposal) error {

	// check if the proposal is for the current round
	round := pro.state.Round()
	if proposal.Block.Height < round {
		return fmt.Errorf("invalid proposal height (proposal: %d, round: %d)", proposal.Block.Height, round)
	}

	// check if the proposal is by correct leader
	leaderID := pro.state.Leader(round)
	if proposal.Block.SignerID != leaderID {
		return fmt.Errorf("invalid proposal signer (proposal: %x, round: %x)", proposal.Block.SignerID, leaderID)
	}

	// check if the proposed block is valid
	valid, err := pro.verify.Proposal(proposal)
	if err != nil {
		return fmt.Errorf("could not verify proposal: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid proposal signature (signer: %x)", proposal.Block.SignerID)
	}

	// store the block in our database
	err = pro.db.Store(proposal.Block)
	if err != nil {
		return fmt.Errorf("could not store block: %w", err)
	}

	// check if we have the proposal parent
	_, err = pro.db.Block(proposal.QC.BlockID)
	if err != nil {
		return fmt.Errorf("could not get proposal parent: %w", err)
	}

	// NOTE: we never check if the QC is on a block that is a valid extension of
	// the state; if it wasn't, the system would already be compromised, because
	// we have a majority of validators voting for an invalid block - we can
	// therefore simply jump to the height after the QC/parent

	// clear the buffer for the voted block
	pro.buf.Clear(proposal.QC.BlockID)

	// set our state to the new height
	pro.state.Set(proposal.Block.Height)

	// check if we were the proposer
	if proposal.Block.SignerID == pro.sign.Self() {
		return nil
	}

	// TODO: check if the proposed block is a valid extension of the state

	// NOTE: if we are the next leader, the network will short-circuit the vote
	// here, and the `OnVote` call will trigger us sending the proposal if we
	// can already assemble enough signatures for the previous one
	// vote on the new proposal

	// create the vote for the proposed block
	vote, err := pro.sign.Vote(proposal.Block)
	if err != nil {
		return fmt.Errorf("could not create vote: %w", err)
	}

	// send the vote for the proposed block to the next collector
	collectorID := pro.state.Leader(proposal.Block.Height + 1)
	err = pro.net.Transmit(vote, collectorID)
	if err != nil {
		return fmt.Errorf("could not transmit vote: %w", err)
	}

	return nil
}

func (pro *Processor) OnVote(vote *message.Vote) error {

	// check if we have the vote block
	candidate, err := pro.db.Block(vote.BlockID)
	if err != nil {
		return fmt.Errorf("could not get vote candidate: %w", err)
	}

	// check if the vote is outdated
	round := pro.state.Round()
	if candidate.Height < round {
		return fmt.Errorf("invalid candidate height (vote: %d, round: %d)", candidate.Height, round)
	}

	// check if we are the collector for the round
	collectorID := pro.state.Leader(candidate.Height + 1)
	if collectorID != pro.sign.Self() {
		return fmt.Errorf("invalid vote collector (collector: %x, self: %x)", collectorID, pro.sign.Self())
	}

	// check if the vote has a valid signature
	valid, err := pro.verify.Vote(vote)
	if err != nil {
		return fmt.Errorf("could not verify vote signature: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid vote signature (signer: %x)", vote.SignerID)
	}

	// check if we already have a vote by this voter
	err = pro.buf.Tally(vote)
	if err != nil {
		return fmt.Errorf("could not tally vote: %w)", err)
	}

	// check to see if we reached threshold
	threshold := pro.state.Threshold()
	votes := pro.buf.Votes(vote.BlockID)
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

	// create the QC for the new proposal
	qc := model.QC{
		BlockID:   vote.BlockID,
		SignerIDs: signerIDs,
		Signature: signature,
	}

	// create the block for the new proposal
	block := model.Block{
		Height:      candidate.Height + 1,
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
