package message

import (
	"github.com/alvalor/consensus/model"
)

type Vote struct {
	Height    uint64
	VertexID  model.Hash
	SignerID  model.Hash
	Signature model.Signature
}
