package save

import "github.com/gentaman/myrogue/internal/core/component"

const SchemaVersion = 1

type Snapshot struct {
	SchemaVersion int              `json:"schema_version"`
	Floor         int              `json:"floor"`
	TurnCount     int              `json:"turn_count"`
	MapSeed       int64            `json:"map_seed"`
	PlayerID      int64            `json:"player_id"`
	Entities      []EntitySnapshot `json:"entities"`
	MapItems      []MapItemSnap    `json:"map_items"`
	MessageLog    []string         `json:"message_log"`
	Explored      []ExploredTile   `json:"explored"`
}

type EntitySnapshot struct {
	ID        int64                 `json:"id"`
	DefID     string                `json:"def_id"`
	Kind      string                `json:"kind"`
	Name      string                `json:"name"`
	Position  *PositionSnap         `json:"position,omitempty"`
	Stats     *StatsSnap            `json:"stats,omitempty"`
	Faction   string                `json:"faction,omitempty"`
	AI        *AISnap               `json:"ai,omitempty"`
	Inventory []component.ItemEntry `json:"inventory,omitempty"`
	Element   string                `json:"element,omitempty"`
	Race      string                `json:"race,omitempty"`
	ColorHex  string                `json:"color_hex,omitempty"`
	HasSprite bool                  `json:"has_sprite,omitempty"`
	RewardXP  int                   `json:"reward_xp,omitempty"`
}

type PositionSnap struct {
	X   int `json:"x"`
	Y   int `json:"y"`
	Dir int `json:"dir"`
}

type StatsSnap struct {
	HP       int `json:"hp"`
	MaxHP    int `json:"max_hp"`
	MP       int `json:"mp"`
	MaxMP    int `json:"max_mp"`
	Level    int `json:"level"`
	XP       int `json:"xp"`
	XPToNext int `json:"xp_to_next"`
	Str      int `json:"str"`
	Wis      int `json:"wis"`
	Fai      int `json:"fai"`
	Vit      int `json:"vit"`
	Agi      int `json:"agi"`
	Luk      int `json:"luk"`
}

type AISnap struct {
	Personality       int  `json:"personality"`
	Order             int  `json:"order"`
	State             int  `json:"state"`
	FriendlyFire      bool `json:"friendly_fire"`
	NeutralThreshold  int  `json:"neutral_threshold"`
	FriendlyThreshold int  `json:"friendly_threshold"`
}

type MapItemSnap struct {
	X     int                   `json:"x"`
	Y     int                   `json:"y"`
	Items []component.ItemEntry `json:"items"`
}

type ExploredTile struct {
	X int `json:"x"`
	Y int `json:"y"`
}
