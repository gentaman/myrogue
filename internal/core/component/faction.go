package component

import "github.com/gentaman/myrogue/internal/core/entity"

type Faction int

const (
	FactionPlayer Faction = iota
	FactionAlly
	FactionEnemy
)

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
