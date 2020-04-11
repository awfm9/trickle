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

func AtRound(height uint64) Condition {
	return func(p *Participant) bool {
		return p.state.Round() >= height
	}
}

func Error() Condition {
	return func(p *Participant) bool {
		return p.last != nil
	}
}
