package message

import (
	"github.com/alvalor/consensus/model"
)

type Vote struct {
	Height    uint64
	BlockID   model.Hash
	SignerID  model.Hash
	Signature []byte
}
