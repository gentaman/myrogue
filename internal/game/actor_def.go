package game

import (
	"encoding/json"
	"fmt"
	"image/color"
)

type Personality int

const (
	PersonalityAggressive Personality = iota
	PersonalityCowardly
	PersonalityCalculated
)

type ActorDef struct {
	ID                           string
	Name                         string
	HP                           int
	MP                           int
	MaxCarryWeight               int
	Element                      Element
	Race                         Race
	Personality                  Personality
	NeutralThreshold             int
	FriendlyThreshold            int
	XP                           int // エネミーの場合は討伐経験値
	Rarity                       int
	Color                        color.RGBA
	Str, Wis, Fai, Vit, Agi, Luk int
}

type rawActor struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	HP                int    `json:"hp"`
	MP                int    `json:"mp"`
	MaxCarryWeight    int    `json:"max_carry_weight"`
	Element           string `json:"element"`
	Race              string `json:"race"`
	Personality       string `json:"personality"`
	NeutralThreshold  int    `json:"neutral_threshold"`
	FriendlyThreshold int    `json:"friendly_threshold"`
	XP                int    `json:"xp"`
	Rarity            int    `json:"rarity"`
	Color             string `json:"color"`
	Str               int    `json:"str"`
	Wis               int    `json:"wis"`
	Fai               int    `json:"fai"`
	Vit               int    `json:"vit"`
	Agi               int    `json:"agi"`
	Luk               int    `json:"luk"`
}

var (
	playerDefs []ActorDef
	enemyDefs  []ActorDef

	playerIDMap = map[string]int{}
	enemyIDMap  = map[string]int{}

	// ショートカット用
	playerMaxHP    int
	maxCarryWeight int
)

func loadActorDefs(data []byte) ([]ActorDef, map[string]int) {
	var raws []rawActor
	if err := json.Unmarshal(data, &raws); err != nil {
		panic(fmt.Sprintf("failed to unmarshal actor data: %v", err))
	}

	defs := make([]ActorDef, len(raws))
	idMap := make(map[string]int)

	for i, raw := range raws {
		idMap[raw.ID] = i
		defs[i] = ActorDef{
			ID:                raw.ID,
			Name:              raw.Name,
			HP:                raw.HP,
			MP:                raw.MP,
			MaxCarryWeight:    raw.MaxCarryWeight,
			Element:           stringToElement(raw.Element),
			Race:              stringToRace(raw.Race),
			Personality:       stringToPersonality(raw.Personality),
			NeutralThreshold:  raw.NeutralThreshold,
			FriendlyThreshold: raw.FriendlyThreshold,
			XP:                raw.XP,
			Rarity:            raw.Rarity,
			Color:             hexToRGBA(raw.Color),
			Str:               raw.Str,
			Wis:               raw.Wis,
			Fai:               raw.Fai,
			Vit:               raw.Vit,
			Agi:               raw.Agi,
			Luk:               raw.Luk,
		}
	}
	return defs, idMap
}

func initActorDefs() {
	playerDefs, playerIDMap = loadActorDefs(playerJSON)
	enemyDefs, enemyIDMap = loadActorDefs(enemiesJSON)

	if len(playerDefs) > 0 {
		playerMaxHP = playerDefs[0].HP
		maxCarryWeight = playerDefs[0].MaxCarryWeight
	}
}

func stringToElement(s string) Element {
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

func stringToRace(s string) Race {
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

func stringToPersonality(s string) Personality {
	switch s {
	case "cowardly":
		return PersonalityCowardly
	case "calculated":
		return PersonalityCalculated
	default:
		return PersonalityAggressive
	}
}
