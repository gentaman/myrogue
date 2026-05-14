package component

type Personality int

const (
	PersonalityAggressive Personality = iota
	PersonalityCowardly
	PersonalityCalculated
)

func PersonalityFromString(s string) Personality {
	switch s {
	case "cowardly":
		return PersonalityCowardly
	case "calculated":
		return PersonalityCalculated
	default:
		return PersonalityAggressive
	}
}

type CompanionOrder int

const (
	OrderFollow CompanionOrder = iota
	OrderWait
	OrderAggressive
)

var CompanionOrderNames = map[CompanionOrder]string{
	OrderFollow:     "ついてきて",
	OrderWait:       "待ってて",
	OrderAggressive: "積極的に戦って",
}

type AlertState int

const (
	AlertIdle AlertState = iota
	AlertAlerted
)

type AI struct {
	Personality       Personality
	Order             CompanionOrder
	State             AlertState
	FriendlyFire      bool
	NeutralThreshold  int
	FriendlyThreshold int
}
