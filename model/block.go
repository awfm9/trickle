package model

import (
	"encoding/json"
	"time"

	"golang.org/x/crypto/sha3"
)

type Block struct {
	Height      uint64
	QC          *QC
	PayloadHash Hash
	Timestamp   time.Time
	SignerID    Hash
}

func (b Block) ID() Hash {
	data, _ := json.Marshal(b)
	hash := sha3.Sum256(data)
	return hash
}
