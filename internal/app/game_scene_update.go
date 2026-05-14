package app

import (
	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/ai"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
	"github.com/gentaman/myrogue/internal/core/turn"
	"github.com/gentaman/myrogue/internal/core/world"
)

func (g *GameScene) isAnyAnimating() bool {
	anyAnim := false
	g.Anims.Each(func(id entity.ID, a *component.AnimState) {
		if a.AttackAnim > 0 || a.DamageAnim > 0 {
			anyAnim = true
		}
	})
	return anyAnim || g.AnimQueue.IsPlaying()
}

func (g *GameScene) tickAnimations() {
	delta := g.Clock.Delta()
	if delta <= 0 {
		return
	}
	g.Anims.Each(func(id entity.ID, a *component.AnimState) {
		if a.AttackAnim > 0 {
			a.AttackAnim -= delta
			if a.AttackAnim < 0 {
				a.AttackAnim = 0
			}
		}
		if a.DamageAnim > 0 {
			a.DamageAnim -= delta
			if a.DamageAnim < 0 {
				a.DamageAnim = 0
			}
		}
	})
	g.AnimQueue.TickWithDelta(delta)
}

func (g *GameScene) resolveCombat(attackerID, defenderID entity.ID) {
	atkPos := g.GetPosition(attackerID)
	defPos := g.GetPosition(defenderID)
	if atkPos == nil || defPos == nil {
		return
	}
	act := &action.AttackAction{TargetX: defPos.X, TargetY: defPos.Y}
	events := act.Execute(attackerID, g)
	g.processEvents(events)
}

func (g *GameScene) Update(input InputState) Scene {
	g.Frame++
	g.handleDebugInput(input)
	g.Clock.Tick()

	if input.DebugItem {
		return &DebugItemScene{game: g}
	}

	if g.isAnyAnimating() {
		g.tickAnimations()
		return nil
	}

	if g.PlayState == StateWin || g.PlayState == StateDead {
		if input.Restart {
			return NewGameScene(g.registry, g.Audio, g.SaveService)
		}
		if input.Cancel {
			return NewTitleSceneWithDeps(g.registry, g.Audio, g.SaveService)
		}
		return nil
	}

	if g.ConfirmQuit {
		if input.Cancel || input.No {
			g.ConfirmQuit = false
		} else if input.Confirm || input.Yes {
			return NewTitleSceneWithDeps(g.registry, g.Audio, g.SaveService)
		}
		return nil
	}

	if g.ConfirmStair {
		if input.Cancel || input.No {
			g.ConfirmStair = false
		} else if input.Confirm || input.Yes {
			g.ConfirmStair = false
			act := &action.FloorChangeAction{Direction: g.PendingDir}
			g.processEvents(act.Execute(g.Player, g))
		}
		return nil
	}

	switch g.Scheduler.Phase {
	case turn.PhasePlayerInput:
		return g.updatePlayerTurn(input)
	case turn.PhaseCompanionAct:
		g.updateCompanionTurns()
	case turn.PhaseEnemyAct:
		g.updateEnemyTurns()
	}
	return nil
}

func (g *GameScene) updatePlayerTurn(input InputState) Scene {
	if input.Help {
		return &HelpScene{game: g}
	}
	if input.Inventory {
		return &InventoryScene{game: g}
	}
	if input.Log {
		return &LogScene{game: g}
	}
	if input.MapView {
		return &MapViewScene{game: g}
	}
	if input.Skill && len(g.registry.Skills) > 0 {
		return NewSkillScene(g)
	}

	if input.Cancel {
		if g.MenuOpen {
			g.MenuOpen = false
			return nil
		}
		g.ConfirmQuit = true
		return nil
	}

	if input.Menu {
		g.MenuOpen = !g.MenuOpen
		if g.MenuOpen {
			g.MenuCursor = 0
			g.buildMenu()
		}
		return nil
	}

	if g.MenuOpen {
		if input.Up {
			g.MenuCursor--
			if g.MenuCursor < 0 {
				g.MenuCursor = len(g.MenuItems) - 1
			}
		}
		if input.Down {
			g.MenuCursor++
			if g.MenuCursor >= len(g.MenuItems) {
				g.MenuCursor = 0
			}
		}
		if input.Confirm {
			if scene := g.execMenuItem(); scene != nil {
				return scene
			}
		}
		return nil
	}

	if input.DirPressed {
		g.handleMovement(input.Dir)
	}
	return nil
}

func (g *GameScene) handleMovement(dir int) {
	var dx, dy int
	switch dir {
	case 0:
		dy = -1
	case 1:
		dy = 1
	case 2:
		dx = -1
	case 3:
		dx = 1
	}
	pPos := g.playerPos()
	nx, ny := pPos.X+dx, pPos.Y+dy

	occupant := g.UnitAt(nx, ny)
	if occupant != entity.InvalidID {
		g.Scheduler.IncrementTurn()
		faction, _ := g.Factions.Get(occupant)
		if faction != nil && faction.Faction == component.FactionAlly {
			// swap with companion
			g.processEvents([]event.Event{event.EventSwap{Entity1: g.Player, Entity2: occupant}})
		} else {
			// attack
			g.resolveCombat(g.Player, occupant)
			if g.Audio != nil {
				g.Audio.PlaySFX("hit")
			}
		}
	} else {
		// move or change dir
		act := &action.MoveAction{DX: dx, DY: dy}
		events := act.Execute(g.Player, g)
		if len(events) > 0 {
			// Only increment turn if we actually tried to move or change dir toward a valid tile
			// (Wait, even changing dir should probably consume a turn? Or not?)
			// Currently, Roguelikes vary. Let's say it consumes a turn for now.
			g.Scheduler.IncrementTurn()
			g.processEvents(events)
		}
	}
	world.UpdateVisibility(g.World, g.playerPos().X, g.playerPos().Y)
	g.checkTileAfterMove()
	g.Scheduler.StartCompanionPhase()
}

func (g *GameScene) checkTileAfterMove() {
	pPos := g.playerPos()
	tile := g.World.Tiles[pPos.X][pPos.Y]
	switch tile {
	case world.Stairs:
		g.pushMessage("出口の階段だ。ここから脱出できる. ")
	case world.StairsDown:
		g.pushMessage("下り階段だ。下の階へ進める。")
	case world.StairsUp:
		g.pushMessage("上り階段だ。上の階へ戻れる。")
	}
}

func (g *GameScene) updateCompanionTurns() {
	allies := g.EntitiesWithFaction(component.FactionAlly)
	if g.Scheduler.ActiveIdx >= len(allies) {
		g.Scheduler.StartEnemyPhase()
		return
	}
	id := allies[g.Scheduler.ActiveIdx]
	g.runAI(id, &ai.AllyBrain{})
	g.Scheduler.AdvanceIdx()
}

func (g *GameScene) updateEnemyTurns() {
	enemies := g.EntitiesWithFaction(component.FactionEnemy)
	if g.Scheduler.ActiveIdx >= len(enemies) {
		g.tickStatusEffects()
		g.Scheduler.StartPlayerPhase()
		g.processEvents([]event.Event{event.EventMP{Entity: g.Player, Amount: 1}})
		return
	}
	id := enemies[g.Scheduler.ActiveIdx]
	g.runAI(id, &ai.HostileBrain{})
	g.Scheduler.AdvanceIdx()
}

func (g *GameScene) runAI(id entity.ID, brain ai.Brain) {
	pos := g.GetPosition(id)
	aiComp := g.GetAI(id)
	if pos == nil {
		return
	}
	if aiComp == nil {
		aiComp = &component.AI{}
	}

	ctx := &ai.DecisionContext{
		Self:    id,
		SelfPos: pos,
		SelfAI:  aiComp,
		World:   g.World,
		Query:   g,
	}
	act := brain.Decide(ctx)
	if act == nil {
		return
	}
	switch a := act.(type) {
	case *action.MoveAction:
		events := a.Execute(id, g)
		g.processEvents(events)
	case *action.AttackAction:
		events := a.Execute(id, g)
		g.processEvents(events)
	case *action.WaitAction:
		// do nothing
	}
}
