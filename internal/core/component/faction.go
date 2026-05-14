package component

import "github.com/gentaman/myrogue/internal/core/entity"

type Faction int

const (
	FactionPlayer Faction = iota
	FactionAlly
	FactionEnemy
)

func (f Faction) String() string {
	switch f {
	case FactionPlayer:
		return "player"
	case FactionAlly:
		return "ally"
	case FactionEnemy:
		return "enemy"
	default:
		return "player"
	}
}

func FactionFromString(s string) Faction {
	switch s {
	case "ally":
		return FactionAlly
	case "enemy":
		return FactionEnemy
	default:
		return FactionPlayer
	}
}

type FactionComp struct {
	Faction   Faction
	Relations map[entity.ID]int
}

func NewFactionComp(f Faction) *FactionComp {
	return &FactionComp{
		Faction:   f,
		Relations: make(map[entity.ID]int),
	}
}

type RelationState int

const (
	RelationHostile RelationState = iota
	RelationNeutral
	RelationFriendly
)
