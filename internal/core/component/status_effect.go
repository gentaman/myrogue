package component

type StatusType int

const (
	StatusPoison StatusType = iota
	StatusParalyze
	StatusConfuse
	StatusBlind
	StatusSleep
	StatusRegen
	StatusAtkUp
	StatusDefUp
)

var StatusNames = map[StatusType]string{
	StatusPoison:   "毒",
	StatusParalyze: "麻痺",
	StatusConfuse:  "混乱",
	StatusBlind:    "暗闇",
	StatusSleep:    "睡眠",
	StatusRegen:    "再生",
	StatusAtkUp:    "攻撃UP",
	StatusDefUp:    "防御UP",
}

func StatusFromString(s string) StatusType {
	switch s {
	case "poison":
		return StatusPoison
	case "paralyze":
		return StatusParalyze
	case "confuse":
		return StatusConfuse
	case "blind":
		return StatusBlind
	case "sleep":
		return StatusSleep
	case "regen":
		return StatusRegen
	case "atk_up":
		return StatusAtkUp
	case "def_up":
		return StatusDefUp
	default:
		return StatusPoison
	}
}

func (s StatusType) String() string {
	switch s {
	case StatusPoison:
		return "poison"
	case StatusParalyze:
		return "paralyze"
	case StatusConfuse:
		return "confuse"
	case StatusBlind:
		return "blind"
	case StatusSleep:
		return "sleep"
	case StatusRegen:
		return "regen"
	case StatusAtkUp:
		return "atk_up"
	case StatusDefUp:
		return "def_up"
	default:
		return "poison"
	}
}

type StatusEffect struct {
	Type     StatusType
	Duration int
}

type StatusEffects struct {
	Effects []StatusEffect
}

func (se *StatusEffects) Has(t StatusType) bool {
	for _, e := range se.Effects {
		if e.Type == t {
			return true
		}
	}
	return false
}

func (se *StatusEffects) Add(t StatusType, duration int) {
	for i, e := range se.Effects {
		if e.Type == t {
			if duration > e.Duration {
				se.Effects[i].Duration = duration
			}
			return
		}
	}
	se.Effects = append(se.Effects, StatusEffect{Type: t, Duration: duration})
}

func (se *StatusEffects) Remove(t StatusType) {
	for i, e := range se.Effects {
		if e.Type == t {
			se.Effects = append(se.Effects[:i], se.Effects[i+1:]...)
			return
		}
	}
}

func (se *StatusEffects) Tick() []StatusType {
	var expired []StatusType
	alive := se.Effects[:0]
	for _, e := range se.Effects {
		e.Duration--
		if e.Duration <= 0 {
			expired = append(expired, e.Type)
		} else {
			alive = append(alive, e)
		}
	}
	se.Effects = alive
	return expired
}
