package message

import (
	"github.com/alvalor/consensus/model"
)

type Proposal struct {
	*model.Block
	Signature []byte
}
