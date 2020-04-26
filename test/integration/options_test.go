package integration

import (
	"github.com/alvalor/consensus/model/base"
	"github.com/rs/zerolog"
)

type Option func(*Participant)

func WithLog(log zerolog.Logger) Option {
	return func(p *Participant) {
		p.log = log
	}
}

func WithSelf(selfID base.Hash) Option {
	return func(p *Participant) {
		p.selfID = selfID
	}
}

func WithParticipants(participantIDs []base.Hash) Option {
	return func(p *Participant) {
		p.participantIDs = participantIDs
	}
}

func WithGenesis(genesisID base.Hash) Option {
	return func(p *Participant) {
		p.genesisID = genesisID
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
