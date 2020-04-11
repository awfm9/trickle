package consensus

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
	"golang.org/x/crypto/sha3"
)

type Processor struct {
	db     Database
	net    Network
	state  State
	sign   Signer
	verify Verifier
	self   model.Hash
	votes  map[model.Hash]*message.Vote
}

func NewProcessor(db Database, net Network, state State, sign Signer, verify Verifier, self model.Hash) *Processor {

	pro := Processor{
		db:     db,
		net:    net,
		state:  state,
		sign:   sign,
		verify: verify,
		self:   self,
		votes:  make(map[model.Hash]*message.Vote),
	}

	return &pro
}

func (pro *Processor) OnVote(vote *message.Vote) error {

	// check if we have the vote block
	block, err := pro.db.Block(vote.BlockID)
	if err != nil {
		return fmt.Errorf("could not get vote block: %w", err)
	}

	// check if the vote is outdated
	height := pro.state.Height()
	if block.Height < height {
		return fmt.Errorf("invalid vote height (vote: %d, round: %d)", block.Height, height)
	}

	// check if we are the collector for the round
	collectorID := pro.state.Leader(block.Height + 1)
	if collectorID != pro.self {
		return fmt.Errorf("invalid vote collector (collector: %x, self: %x)", collectorID, pro.self)
	}

	// check if we already have a vote by this voter
	_, voted := pro.votes[vote.SignerID]
	if voted {
		return fmt.Errorf("signer has already voted (signer: %x)", vote.SignerID)
	}

	// check if the vote has a valid signature
	valid, err := pro.verify.Vote(vote)
	if err != nil {
		return fmt.Errorf("could not verify vote signature: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid vote signature (signer: %x)", vote.SignerID)
	}

	// add the vote to our buffer
	pro.votes[vote.SignerID] = vote

	// check to see if we reached threshold
	threshold := len(pro.state.Participants()) / 3 * 2
	if len(pro.votes) <= threshold {
		return nil
	}

	// collect vote signers and signatures
	var signature []byte
	signerIDs := make([]model.Hash, 0, len(pro.votes))
	for signerID, vote := range pro.votes {
		signerIDs = append(signerIDs, signerID)
		signature = append(signature, vote.Signature...)
	}

	// get a random payload
	payload := make([]byte, 1024)
	_, _ = rand.Read(payload)

	// create the QC for the new proposal
	qc := model.QC{
		BlockID:   vote.BlockID,
		SignerIDs: signerIDs,
		Signature: signature,
	}

	// create the block for the new proposal
	candidate := model.Block{
		Height:      block.Height + 1,
		QC:          &qc,
		PayloadHash: sha3.Sum256(payload),
		Timestamp:   time.Now().UTC(),
		SignerID:    pro.self,
	}

	// create the new proposal
	proposal, err := pro.sign.Proposal(&candidate)
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

func (pro *Processor) OnProposal(proposal *message.Proposal) error {

	// check if the proposal is for the current round
	height := pro.state.Height()
	if proposal.Block.Height < height {
		return fmt.Errorf("invalid proposal height (proposal: %d, round: %d)", proposal.Block.Height, height)
	}

	// check if the proposal is by correct leader
	leaderID := pro.state.Leader(height)
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

	// check if we have the proposal parent
	_, err = pro.db.Block(proposal.QC.BlockID)
	if err != nil {
		return fmt.Errorf("could not get proposal parent: %w", err)
	}

	// NOTE: we never check if the QC is on a block that is a valid extension of
	// the state; if it wasn't, the system would already be compromised, because
	// we have a majority of validators voting for an invalid block - we can
	// therefore simply jump to the height after the QC/parent

	// empty the map of votes; we iterate as most of the time we are not the
	// collector, so it is empty anyway, no need to throw away
	for signerID := range pro.votes {
		delete(pro.votes, signerID)
	}

	// set our state to the new height
	pro.state.Set(proposal.Block.Height)

	// check if we were the proposer
	if proposal.Block.SignerID == pro.self {
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
