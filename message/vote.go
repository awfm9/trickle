package message

import (
	"github.com/alvalor/consensus/model"
)

type Vote struct {
	BlockID   model.Hash
	SignerID  model.Hash
	Signature []byte
}
