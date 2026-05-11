package ai

import (
	"math/rand"

	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

type AllyBrain struct{}

func (b *AllyBrain) Decide(ctx *DecisionContext) action.Action {
	if ctx.SelfAI.Order == component.OrderWait {
		return &action.WaitAction{}
	}

	target, tx, ty := b.findHostile(ctx)
	if target != entity.InvalidID {
		dx := tx - ctx.SelfPos.X
		dy := ty - ctx.SelfPos.Y
		if world.Abs(dx)+world.Abs(dy) == 1 {
			if ctx.SelfAI.FriendlyFire && b.friendlyFireCheck(ctx, tx, ty) {
				return &action.WaitAction{}
			}
			return &action.AttackAction{TargetX: tx, TargetY: ty}
		}
		return b.moveToward(ctx, tx, ty)
	}

	playerPos := ctx.Query.GetPosition(ctx.Query.PlayerID())
	if playerPos == nil {
		return &action.WaitAction{}
	}
	dist := world.Abs(ctx.SelfPos.X-playerPos.X) + world.Abs(ctx.SelfPos.Y-playerPos.Y)
	if dist <= 1 {
		return &action.WaitAction{}
	}
	return b.moveToward(ctx, playerPos.X, playerPos.Y)
}

func (b *AllyBrain) findHostile(ctx *DecisionContext) (entity.ID, int, int) {
	enemies := ctx.Query.EntitiesWithFaction(component.FactionEnemy)
	for _, id := range enemies {
		pos := ctx.Query.GetPosition(id)
		if pos == nil || !ctx.Query.IsAlive(id) {
			continue
		}
		if world.CanSee(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, pos.X, pos.Y) {
			rel := b.getRelation(ctx, id)
			if rel == component.RelationHostile {
				return id, pos.X, pos.Y
			}
		}
	}
	return entity.InvalidID, 0, 0
}

func (b *AllyBrain) getRelation(ctx *DecisionContext, target entity.ID) component.RelationState {
	selfFaction := ctx.Query.GetFaction(ctx.Self)
	if selfFaction == nil {
		return component.RelationNeutral
	}
	score := selfFaction.Relations[target]
	if score <= ctx.SelfAI.NeutralThreshold {
		return component.RelationHostile
	}
	if score >= ctx.SelfAI.FriendlyThreshold {
		return component.RelationFriendly
	}
	if ctx.SelfAI.Order == component.OrderAggressive {
		return component.RelationHostile
	}
	return component.RelationNeutral
}

func (b *AllyBrain) moveToward(ctx *DecisionContext, tx, ty int) action.Action {
	blocked := func(x, y int) bool {
		return ctx.Query.UnitAt(x, y) != entity.InvalidID
	}
	nx, ny := world.BFSNextStep(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, tx, ty, blocked)
	if nx == ctx.SelfPos.X && ny == ctx.SelfPos.Y {
		return &action.WaitAction{}
	}
	if ctx.Query.UnitAt(nx, ny) != entity.InvalidID {
		return &action.WaitAction{}
	}
	return &action.MoveAction{DX: nx - ctx.SelfPos.X, DY: ny - ctx.SelfPos.Y}
}

func (b *AllyBrain) friendlyFireCheck(ctx *DecisionContext, targetX, targetY int) bool {
	dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	hasFriendly := false
	for _, d := range dirs {
		ax, ay := targetX+d[0], targetY+d[1]
		if ax == ctx.SelfPos.X && ay == ctx.SelfPos.Y {
			continue
		}
		occupant := ctx.Query.UnitAt(ax, ay)
		if occupant == entity.InvalidID {
			continue
		}
		faction := ctx.Query.GetFaction(occupant)
		if faction != nil && (faction.Faction == component.FactionPlayer || faction.Faction == component.FactionAlly) {
			hasFriendly = true
			break
		}
	}
	if !hasFriendly {
		return false
	}
	return rand.Intn(100) < 15
}
