package consensus

import (
	"github.com/alvalor/consensus/message"
)

type Verifier interface {
	Proposal(proposal *message.Proposal) error
	Vote(vote *message.Vote) error
}
