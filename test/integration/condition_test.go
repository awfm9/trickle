package integration

import (
	"time"
)

type Condition func(*Participant) error

func AfterDelay(delay time.Duration, err error) Condition {
	start := time.Now()
	return func(p *Participant) error {
		if time.Now().After(start.Add(delay)) {
			return err
		}
		return nil
	}
}
