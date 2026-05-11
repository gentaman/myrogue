package component

type Element int

const (
	ElementNone Element = iota
	ElementFire
	ElementWater
	ElementAir
	ElementEarth
	ElementLight
	ElementDark
)

var ElementEffectiveness = map[Element]map[Element]float64{
	ElementFire:  {ElementAir: 2.0},
	ElementWater: {ElementFire: 2.0},
	ElementAir:   {ElementEarth: 2.0},
	ElementEarth: {ElementWater: 2.0},
	ElementLight: {ElementDark: 2.0},
	ElementDark:  {ElementLight: 2.0},
}

func GetEffectiveness(attacker, defender Element) float64 {
	if defs, ok := ElementEffectiveness[attacker]; ok {
		if rate, ok := defs[defender]; ok {
			return rate
		}
	}
	return 1.0
}

func ElementFromString(s string) Element {
	switch s {
	case "fire":
		return ElementFire
	case "water":
		return ElementWater
	case "air":
		return ElementAir
	case "earth":
		return ElementEarth
	case "light":
		return ElementLight
	case "dark":
		return ElementDark
	default:
		return ElementNone
	}
}
