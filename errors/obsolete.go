package errors

import (
	"fmt"

	"github.com/alvalor/consensus/message"
	"github.com/alvalor/consensus/model"
)

type ObsoleteProposal struct {
	Vertex *model.Vertex
	Cutoff uint64
}

func (op ObsoleteProposal) Error() string {
	return fmt.Sprintf("obsolete proposal (height: %d, cutoff: %d)", op.Vertex.Height, op.Cutoff)
}

type ObsoleteVote struct {
	Vote   *message.Vote
	Cutoff uint64
}

func (ov ObsoleteVote) Error() string {
	return fmt.Sprintf("obsolete vote (height: %d, cutoff: %d)", ov.Vote.Height, ov.Cutoff)
}

type StaleVote struct {
	Vote *message.Vote
}

func (sv StaleVote) Error() string {
	return fmt.Sprintf("stale vote (height: %d)", sv.Vote.Height)
}
