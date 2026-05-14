package content

import "github.com/gentaman/myrogue/internal/core/component"

type EffectSpec struct {
	Type      string `json:"type"`
	Amount    int    `json:"amount"`
	Message   string `json:"message"`
	EquipType string `json:"equip_type"`
	Atk       int    `json:"atk"`
	Def       int    `json:"def"`
	Cost      int    `json:"cost"`
	Damage    int    `json:"damage"`
	Range     int    `json:"range"`
	Color     string `json:"color"`
	Element   string `json:"element"`
}

type ConditionSpec struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ItemDef struct {
	ID         string
	Name       string
	Desc       string
	EffectDesc string
	Weight     int
	Rarity     int
	ColorHex   string
	Durability int
	FloorBound bool
	EquipSlot  component.EquipSlot
	PhyAtk     int
	PhyDef     int
	MagAtk     int
	MagDef     int
	Effect     *EffectSpec
	Condition  *ConditionSpec
}

type rawItem struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Desc       string         `json:"desc"`
	EffectDesc string         `json:"effect_desc"`
	Weight     int            `json:"weight"`
	Rarity     int            `json:"rarity"`
	Color      string         `json:"color"`
	Durability int            `json:"durability"`
	FloorBound bool           `json:"floor_bound"`
	CanUse     *ConditionSpec `json:"can_use"`
	Effect     *EffectSpec    `json:"effect"`
}
