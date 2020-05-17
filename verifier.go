package consensus

import (
	"github.com/awfm/consensus/model/message"
)

type Verifier interface {
	Quorum(quorum *message.Proposal) error
	Proposal(proposal *message.Proposal) error
	Vote(vote *message.Vote) error
}
