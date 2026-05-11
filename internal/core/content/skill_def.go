package content

import "github.com/gentaman/myrogue/internal/core/component"

type SkillDef struct {
	ID       string
	Name     string
	Desc     string
	Type     SkillType
	Power    int
	MPCost   int
	Range    int
	Element  component.Element
	Status   string
	Duration int
	ColorHex string
}

type SkillType int

const (
	SkillTypeAttack SkillType = iota
	SkillTypeHeal
	SkillTypeBuff
	SkillTypeDebuff
)

type rawSkill struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Type     string `json:"type"`
	Power    int    `json:"power"`
	MPCost   int    `json:"mp_cost"`
	Range    int    `json:"range"`
	Element  string `json:"element"`
	Status   string `json:"status"`
	Duration int    `json:"duration"`
	Color    string `json:"color"`
}

func skillTypeFromString(s string) SkillType {
	switch s {
	case "heal":
		return SkillTypeHeal
	case "buff":
		return SkillTypeBuff
	case "debuff":
		return SkillTypeDebuff
	default:
		return SkillTypeAttack
	}
}
