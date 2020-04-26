package signal

import (
	"fmt"

	"github.com/alvalor/consensus/model/base"
	"github.com/alvalor/consensus/model/message"
)

// StaleVote is an error returned when processing of a vote is skipped because
// the vote was cast for a candidate that has already been included in our graph
// state.
type StaleVote struct {
	Vote *message.Vote
}

func (sv StaleVote) Error() string {
	return fmt.Sprintf("stale vote (height: %d, candidate: %x)", sv.Vote.Height, sv.Vote.CandidateID)
}

// ConflictingVote is an error returned when processing of a vote is skipped
// because the vote is for a condidate that is in conflict with the immutable
// state.
type ConflictingVote struct {
	Vote  *message.Vote
	Final *base.Vertex
}

func (cv ConflictingVote) Error() string {
	return fmt.Sprintf("conflicting vote (height: %d, final: %d)", cv.Vote.Height, cv.Final.Height)
}

// ObsoleteVote is an error returned when processing of a vote is skipped
// because the vote is for a candidate that is already behind another candidate
// with quorum, thus probably never being finalized.
type ObsoleteVote struct {
	Vote *message.Vote
	Tip  *base.Vertex
}

func (ov ObsoleteVote) Error() string {
	return fmt.Sprintf("obsolete vote (height: %d, tip: %d)", ov.Vote.Height, ov.Tip.Height)
}

// InvalidCollector is an error returned when processing a vote that has been
// sent to the wrong collector, and the intended collector is not the recipient
// (which is usually ourselves).
type InvalidCollector struct {
	Vote      *message.Vote
	Receiver  base.Hash
	Collector base.Hash
}

func (ic InvalidCollector) Error() string {
	return fmt.Sprintf("invalid collector (sender: %x, receiver: %x, collector: %x)", ic.Vote.SignerID, ic.Receiver, ic.Collector)
}

// DoubleVote is an error returned when trying to store a vote by a voter who
// has already voted for a different proposal at the same height.
type DoubleVote struct {
	First  *message.Vote
	Second *message.Vote
}

func (dv DoubleVote) Error() string {
	return fmt.Sprintf("double vote (height: %d, voter: %x, candidate1: %x, candidate2: %x)", dv.First.Height, dv.First.SignerID, dv.First.CandidateID, dv.Second.CandidateID)
}
