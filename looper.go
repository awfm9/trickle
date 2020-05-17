package consensus

import (
	"github.com/awfm/consensus/model/message"
)

// Looper is used to loop back messages to ourselves, with priority, thus
// pre-empting other messages that might be submitted next.
type Looper interface {
	Proposal(proposal *message.Proposal)
	Vote(vote *message.Vote)
}
