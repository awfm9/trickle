package integration

type Condition func(*Participant) bool

func Or(conditions ...Condition) Condition {
	return func(p *Participant) bool {
		for _, condition := range conditions {
			if condition(p) {
				return true
			}
		}
		return false
	}
}

func AfterRound(height uint64) Condition {
	return func(p *Participant) bool {
		return p.round > height
	}
}
