package ai

import (
	"math/rand"

	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

type HostileBrain struct{}

func (b *HostileBrain) Decide(ctx *DecisionContext) action.Action {
	target, tx, ty := b.findTarget(ctx)
	if target == entity.InvalidID {
		return b.wander(ctx)
	}

	dx := tx - ctx.SelfPos.X
	dy := ty - ctx.SelfPos.Y
	if world.Abs(dx)+world.Abs(dy) == 1 {
		return &action.AttackAction{TargetX: tx, TargetY: ty}
	}

	personality := ctx.SelfAI.Personality
	var nx, ny int
	flee := false
	switch personality {
	case component.PersonalityCowardly:
		flee = true
	case component.PersonalityCalculated:
		playerPos := ctx.Query.GetPosition(ctx.Query.PlayerID())
		if playerPos != nil {
			stats := ctx.Query.GetFaction(ctx.Query.PlayerID())
			_ = stats
			flee = false
		}
	}

	blocked := func(x, y int) bool {
		return ctx.Query.UnitAt(x, y) != entity.InvalidID
	}

	if flee {
		nx, ny = world.BFSFleeStep(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, tx, ty, blocked)
	} else {
		nx, ny = world.BFSNextStep(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, tx, ty, blocked)
	}

	if nx == ctx.SelfPos.X && ny == ctx.SelfPos.Y {
		return &action.WaitAction{}
	}
	if ctx.Query.UnitAt(nx, ny) != entity.InvalidID {
		return &action.WaitAction{}
	}
	return &action.MoveAction{DX: nx - ctx.SelfPos.X, DY: ny - ctx.SelfPos.Y}
}

func (b *HostileBrain) findTarget(ctx *DecisionContext) (entity.ID, int, int) {
	playerID := ctx.Query.PlayerID()
	playerPos := ctx.Query.GetPosition(playerID)
	if playerPos != nil && world.CanSee(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, playerPos.X, playerPos.Y) {
		selfFaction := ctx.Query.GetFaction(ctx.Self)
		if selfFaction != nil {
			rel := getRelation(ctx.Query, ctx.Self, playerID)
			if rel == component.RelationHostile {
				return playerID, playerPos.X, playerPos.Y
			}
		}
	}

	allies := ctx.Query.EntitiesWithFaction(component.FactionAlly)
	for _, id := range allies {
		pos := ctx.Query.GetPosition(id)
		if pos == nil || !ctx.Query.IsAlive(id) {
			continue
		}
		if world.CanSee(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, pos.X, pos.Y) {
			rel := getRelation(ctx.Query, ctx.Self, id)
			if rel == component.RelationHostile {
				return id, pos.X, pos.Y
			}
		}
	}

	return entity.InvalidID, 0, 0
}

func (b *HostileBrain) wander(ctx *DecisionContext) action.Action {
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rand.Shuffle(len(dirs), func(a, b int) { dirs[a], dirs[b] = dirs[b], dirs[a] })
	for _, d := range dirs {
		nx, ny := ctx.SelfPos.X+d[0], ctx.SelfPos.Y+d[1]
		if ctx.World.IsPassable(nx, ny) && ctx.Query.UnitAt(nx, ny) == entity.InvalidID {
			return &action.MoveAction{DX: d[0], DY: d[1]}
		}
	}
	return &action.WaitAction{}
}

func getRelation(q EntityQuery, self, target entity.ID) component.RelationState {
	selfFaction := q.GetFaction(self)
	targetFaction := q.GetFaction(target)
	if selfFaction == nil || targetFaction == nil {
		return component.RelationNeutral
	}

	selfRace := q.GetRace(self)
	targetRace := q.GetRace(target)
	if selfRace == targetRace {
		return component.RelationFriendly
	}

	score := selfFaction.Relations[target]
	selfAI := q.GetAI(self)
	if selfAI != nil {
		if score <= selfAI.NeutralThreshold {
			return component.RelationHostile
		}
		if score >= selfAI.FriendlyThreshold {
			return component.RelationFriendly
		}
	}

	if selfFaction.Faction == component.FactionEnemy {
		return component.RelationHostile
	}
	return component.RelationNeutral
}
