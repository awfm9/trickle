package model

import (
	"encoding/json"

	"golang.org/x/crypto/sha3"
)

type Block struct {
	Height      uint64
	ParentID    Hash
	LeaderID    Hash
	PayloadHash Hash
}

func (b Block) ID() Hash {
	data, _ := json.Marshal(b)
	hash := sha3.Sum256(data)
	return hash
}
