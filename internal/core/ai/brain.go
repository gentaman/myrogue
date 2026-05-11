package ai

import (
	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

type EntityQuery interface {
	EntitiesWithFaction(f component.Faction) []entity.ID
	GetPosition(id entity.ID) *component.Position
	GetFaction(id entity.ID) *component.FactionComp
	GetAI(id entity.ID) *component.AI
	GetRace(id entity.ID) component.Race
	IsAlive(id entity.ID) bool
	UnitAt(x, y int) entity.ID
	PlayerID() entity.ID
}

type DecisionContext struct {
	Self    entity.ID
	SelfPos *component.Position
	SelfAI  *component.AI
	World   *world.GameMap
	Query   EntityQuery
}

type Brain interface {
	Decide(ctx *DecisionContext) action.Action
}
