package message

import (
	"github.com/alvalor/consensus/model"
)

type Proposal struct {
	*model.Block
	Signature []byte
}

func (p *Proposal) Vote() *Vote {
	vote := Vote{
		BlockID:   p.Block.ID(),
		SignerID:  p.SignerID,
		Signature: p.Signature,
	}
	return &vote
}
