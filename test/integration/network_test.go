package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
)

func Network(t testing.TB, participants ...*Participant) {

	// create a list & a registry for participants
	registry := make(map[base.Hash]*Participant)
	for _, p := range participants {
		registry[p.selfID] = p
	}

	// update the set of participants for each participant and update the
	// network mock to properly connect them
	for i := range participants {
		sender := participants[i]

		*sender.net = mocks.Network{}
		sender.net.On("Broadcast", mock.Anything).Return(
			func(proposal *message.Proposal) error {
				for _, receiver := range participants {
					receiver.proposalQ <- proposal
				}
				candidateID := proposal.Candidate.ID()
				sender.log.Debug().
					Uint64("height", proposal.Candidate.Height).
					Hex("candidate", candidateID[:]).
					Hex("parent", proposal.Candidate.ParentID[:]).
					Hex("arc", proposal.Candidate.ArcID[:]).
					Hex("proposer", proposal.Candidate.ProposerID[:]).
					Msg("proposal broadcasted")
				return nil
			},
		)
		sender.net.On("Transmit", mock.Anything, mock.Anything).Return(
			func(vote *message.Vote, recipientID base.Hash) error {
				receiver, exists := registry[recipientID]
				if !exists {
					return fmt.Errorf("invalid recipient (%x)", recipientID)
				}
				receiver.voteQ <- vote
				sender.log.Debug().
					Uint64("height", vote.Height).
					Hex("candidate", vote.CandidateID[:]).
					Hex("voter", vote.SignerID[:]).
					Msg("vote transmitted")
				return nil
			},
		)
	}
}
