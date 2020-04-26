package message

import (
	"github.com/alvalor/consensus/model/base"
)

// Vote is a vot in favour of a new vertex in the consensus graph. It contains
// the height and the ID of the proposed vertex, as well as the ID and the
// signature of the voter.
type Vote struct {
	Height      uint64
	CandidateID base.Hash
	SignerID    base.Hash
	Signature   base.Signature
}
