package integration

import (
	"fmt"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/mocks"
	"github.com/alvalor/consensus/model"
)

func Network(t require.TestingT, participants ...*Participant) {

	// create a list & a registry for participants
	registry := make(map[model.Hash]*Participant)
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
				vertexID := proposal.Vertex.ID()
				sender.log.Debug().
					Uint64("height", proposal.Height).
					Hex("vertex", vertexID[:]).
					Hex("parent", proposal.Parent.VertexID[:]).
					Hex("arc", proposal.ArcID[:]).
					Hex("proposer", proposal.SignerID[:]).
					Msg("proposal broadcasted")
				return nil
			},
		)
		sender.net.On("Transmit", mock.Anything, mock.Anything).Return(
			func(vote *message.Vote, recipientID model.Hash) error {
				receiver, exists := registry[recipientID]
				if !exists {
					return fmt.Errorf("invalid recipient (%x)", recipientID)
				}
				receiver.voteQ <- vote
				sender.log.Debug().
					Uint64("height", vote.Height).
					Hex("vertex", vote.VertexID[:]).
					Hex("voter", vote.SignerID[:]).
					Msg("vote transmitted")
				return nil
			},
		)
	}
}
