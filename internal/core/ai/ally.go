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
		// スキルの検討
		if act := b.trySkill(ctx, target, tx, ty); act != nil {
			return act
		}

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

func (b *AllyBrain) trySkill(ctx *DecisionContext, target entity.ID, tx, ty int) action.Action {
	skillIDs := ctx.Query.GetSkills(ctx.Self)
	if len(skillIDs) == 0 {
		return nil
	}

	stats := ctx.Query.GetStats(ctx.Self)
	if stats == nil {
		return nil
	}

	// 性格による魔法使用率の決定
	useChance := 20
	switch ctx.SelfAI.Personality {
	case component.PersonalityAggressive:
		useChance = 15
	case component.PersonalityCalculated:
		useChance = 50
	case component.PersonalityCowardly:
		useChance = 70
	}

	if ctx.Query.RNG().Intn(100) >= useChance {
		return nil
	}

	dist := world.Abs(ctx.SelfPos.X-tx) + world.Abs(ctx.SelfPos.Y-ty)

	// 使用可能なスキルを探す
	reg := ctx.Query.Registry()
	for _, id := range skillIDs {
		sk, ok := reg.GetSkillDef(id)
		if !ok || stats.MP < sk.MPCost {
			continue
		}

		if sk.Range >= dist {
			return &action.SkillAction{SkillID: id, TargetX: tx, TargetY: ty}
		}
	}

	return nil
}

func (b *AllyBrain) findHostile(ctx *DecisionContext) (entity.ID, int, int) {
	var bestID entity.ID = entity.InvalidID
	var bx, by int
	minDist := 999

	// 敵陣営のユニットを検索
	enemies := ctx.Query.EntitiesWithFaction(component.FactionEnemy)
	for _, id := range enemies {
		pos := ctx.Query.GetPosition(id)
		if pos == nil || !ctx.Query.IsAlive(id) {
			continue
		}
		dist := world.Abs(ctx.SelfPos.X-pos.X) + world.Abs(ctx.SelfPos.Y-pos.Y)
		if dist < minDist && world.CanSee(ctx.World, ctx.SelfPos.X, ctx.SelfPos.Y, pos.X, pos.Y) {
			rel := b.getRelation(ctx, id)
			if rel == component.RelationHostile {
				minDist = dist
				bestID = id
				bx, by = pos.X, pos.Y
			}
		}
	}

	return bestID, bx, by
}

func (b *AllyBrain) getRelation(ctx *DecisionContext, target entity.ID) component.RelationState {
	selfFaction := ctx.Query.GetFaction(ctx.Self)
	targetFaction := ctx.Query.GetFaction(target)
	if selfFaction == nil || targetFaction == nil {
		return component.RelationNeutral
	}

	// 同じ陣営なら友好的
	if selfFaction.Faction == targetFaction.Faction || targetFaction.Faction == component.FactionPlayer {
		return component.RelationFriendly
	}

	// 敵対陣営なら敵対
	if targetFaction.Faction == component.FactionEnemy {
		return component.RelationHostile
	}

	// 個別の関係性スコアによる判定
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
