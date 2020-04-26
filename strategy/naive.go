package strategy

import (
	"math/rand"

	"github.com/alvalor/consensus/model/base"
)

type Naive struct {
	participantIDs []base.Hash
}

func NewNaive(participantIDs []base.Hash) *Naive {
	n := Naive{
		participantIDs: participantIDs,
	}
	return &n
}

func (n *Naive) Threshold(height uint64) (uint, error) {
	return uint(len(n.participantIDs) * 2 / 3), nil
}

func (n *Naive) Leader(height uint64) (base.Hash, error) {
	src := rand.NewSource(int64(height))
	r := rand.New(src)
	index := r.Intn(len(n.participantIDs))
	return n.participantIDs[index], nil
}

func (n *Naive) Collector(height uint64) (base.Hash, error) {
	src := rand.NewSource(int64(height + 1))
	r := rand.New(src)
	index := r.Intn(len(n.participantIDs))
	return n.participantIDs[index], nil
}
