package consensus

import (
	"github.com/alvalor/consensus/message"
)

type Verifier interface {
	Proposal(proposal *message.Proposal) (bool, error)
	Vote(vote *message.Vote) (bool, error)
}
