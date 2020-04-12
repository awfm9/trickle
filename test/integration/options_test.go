package integration

import (
	"github.com/alvalor/consensus/model"
	"github.com/rs/zerolog"
)

type Option func(*Participant)

func WithLog(log zerolog.Logger) Option {
	return func(p *Participant) {
		p.log = log
	}
}

func WithSelf(selfID model.Hash) Option {
	return func(p *Participant) {
		p.selfID = selfID
	}
}

func WithParticipants(participantIDs []model.Hash) Option {
	return func(p *Participant) {
		p.participantIDs = participantIDs
	}
}

func WithRound(round uint64) Option {
	return func(p *Participant) {
		p.round = round
	}
}

func WithIgnore(errs ...error) Option {
	return func(p *Participant) {
		p.ignore = errs
	}
}

func WithStop(conditions ...Condition) Option {
	return func(p *Participant) {
		p.stop = conditions
	}
}
