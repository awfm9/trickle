package integration

import (
	"time"
)

type Condition func(*Participant) error

func AfterRound(height uint64, err error) Condition {
	return func(p *Participant) error {
		if p.round >= height {
			return err
		}
		return nil
	}
}

func AfterDelay(delay time.Duration, err error) Condition {
	start := time.Now()
	return func(p *Participant) error {
		if time.Now().After(start.Add(delay)) {
			return err
		}
		return nil
	}
}
