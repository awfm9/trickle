package signal

import (
	"fmt"

	"github.com/alvalor/consensus/model/base"
)

// InvalidSignature is an error that is returned when a cryptographically
// invalid signature is attached to a message.
type InvalidSignature struct {
	Entity string
	Signer base.Hash
}

func (is InvalidSignature) Error() string {
	return fmt.Sprintf("invalid signature (entity: %T, signer: %x)", is.Entity, is.Signer)
}
