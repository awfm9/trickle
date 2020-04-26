package message

import (
	"github.com/alvalor/consensus/model/base"
)

// Quorum is a collection of signers and their combined signatures.
type Quorum struct {
	SignerIDs []base.Hash
	Signature base.Signature
}
