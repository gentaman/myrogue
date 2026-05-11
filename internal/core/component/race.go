package component

type Race int

const (
	RaceHuman Race = iota
	RaceElf
	RaceDwarf
	RaceGnome
	RaceHalfling
	RaceElement
	RaceBeast
	RaceDragon
	RacePlant
	RaceUndead
	RaceInsect
	RaceBird
	RaceDemon
	RaceMachine
	RaceHolyBeast
)

var RaceNames = map[Race]string{
	RaceHuman:     "ヒューマン",
	RaceElf:       "エルフ",
	RaceDwarf:     "ドワーフ",
	RaceGnome:     "ノーム",
	RaceHalfling:  "ハーフリング",
	RaceElement:   "エレメント",
	RaceBeast:     "野獣",
	RaceDragon:    "ドラゴン",
	RacePlant:     "植物",
	RaceUndead:    "アンデッド",
	RaceInsect:    "虫",
	RaceBird:      "鳥",
	RaceDemon:     "悪魔",
	RaceMachine:   "機械",
	RaceHolyBeast: "聖獣",
}

var OrderedRaces = []Race{
	RaceHuman, RaceElf, RaceDwarf, RaceGnome, RaceHalfling,
	RaceElement, RaceBeast, RaceDragon, RacePlant, RaceUndead,
	RaceInsect, RaceBird, RaceDemon, RaceMachine, RaceHolyBeast,
}

func RaceFromString(s string) Race {
	switch s {
	case "human":
		return RaceHuman
	case "elf":
		return RaceElf
	case "dwarf":
		return RaceDwarf
	case "gnome":
		return RaceGnome
	case "halfling":
		return RaceHalfling
	case "element":
		return RaceElement
	case "beast":
		return RaceBeast
	case "dragon":
		return RaceDragon
	case "plant":
		return RacePlant
	case "undead":
		return RaceUndead
	case "insect":
		return RaceInsect
	case "bird":
		return RaceBird
	case "demon":
		return RaceDemon
	case "machine":
		return RaceMachine
	case "holy_beast":
		return RaceHolyBeast
	default:
		return RaceHuman
	}
}
