package content

import "github.com/gentaman/myrogue/internal/core/component"

type ActorDef struct {
	ID                string
	Name              string
	HP                int
	MP                int
	MaxCarryWeight    int
	Element           component.Element
	Race              component.Race
	Personality       component.Personality
	NeutralThreshold  int
	FriendlyThreshold int
	FriendlyFire      bool
	XP                int
	Rarity            int
	ColorHex          string
	Str               int
	Wis               int
	Fai               int
	Vit               int
	Agi               int
	Luk               int
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
	FriendlyFire      bool   `json:"friendly_fire"`
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
