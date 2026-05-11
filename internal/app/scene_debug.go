package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/turn"
	"github.com/gentaman/myrogue/internal/debug"
)

func (g *GameScene) handleDebugInput(input InputState) {
	if !debug.Enabled {
		return
	}
	if input.Debug {
		g.Debug.Toggle()
	}
	if !g.Debug.Enabled {
		return
	}
	if input.DebugFOV {
		g.Debug.ShowFOV = !g.Debug.ShowFOV
		// TODO: Implement FOV visualization
		// - Highlight tiles in FOV with distinctive color
		// - Show entities outside FOV when enabled
	}
	if input.DebugGrid {
		g.Debug.ShowGrid = !g.Debug.ShowGrid
	}
	if input.DebugEntityID {
		g.Debug.ShowEntityID = !g.Debug.ShowEntityID
	}
	if input.DebugReveal {
		g.Debug.RevealMap = !g.Debug.RevealMap
	}
	if input.DebugSlowAnim {
		s := g.Clock.Scale()
		if s > 0.25 {
			g.Clock.SetScale(s * 0.5)
		}
	}
	if input.DebugFastAnim {
		s := g.Clock.Scale()
		if s < 4.0 {
			g.Clock.SetScale(s * 2.0)
		}
	}
	if input.DebugPause {
		g.Clock.SetPaused(!g.Clock.Paused())
	}
	if input.DebugStep {
		g.Clock.StepFrame()
	}
	if input.DebugSkipAnim {
		g.AnimQueue.SkipAll()
		g.Anims.Each(func(id entity.ID, a *component.AnimState) {
			a.AttackAnim = 0
			a.DamageAnim = 0
		})
	}
}

func (g *GameScene) debugSnapshot() debug.Snapshot {
	snap := debug.Snapshot{
		Floor:      g.World.Floor + 1,
		Turn:       g.Scheduler.TurnCount,
		AnimSpeed:  g.Clock.Scale(),
		AnimPaused: g.Clock.Paused(),
		AnimQueue:  len(g.AnimQueue.Projectiles),
	}
	switch g.Scheduler.Phase {
	case turn.PhasePlayerInput:
		snap.Phase = "Player"
	case turn.PhaseCompanionAct:
		snap.Phase = "Companion"
	case turn.PhaseEnemyAct:
		snap.Phase = "Enemy"
	}
	snap.EntityCount = g.Entities.Count()
	if stats, ok := g.Stats.Get(g.Player); ok {
		snap.PlayerHP = stats.HP
		snap.PlayerMaxHP = stats.MaxHP
		snap.PlayerMP = stats.MP
		snap.PlayerMaxMP = stats.MaxMP
	}
	if pos, ok := g.Positions.Get(g.Player); ok {
		snap.PlayerX = pos.X
		snap.PlayerY = pos.Y
	}
	return snap
}

func (g *GameScene) drawDebugHUD(r Renderer) {
	if !debug.Enabled || !g.Debug.ShowHUD {
		return
	}
	snap := g.debugSnapshot()
	lines := []string{
		fmt.Sprintf("Turn:%d Phase:%s", snap.Turn, snap.Phase),
		fmt.Sprintf("Entities:%d", snap.EntityCount),
		fmt.Sprintf("Player HP:%d/%d MP:%d/%d", snap.PlayerHP, snap.PlayerMaxHP, snap.PlayerMP, snap.PlayerMaxMP),
		fmt.Sprintf("Pos:(%d,%d)", snap.PlayerX, snap.PlayerY),
		fmt.Sprintf("Anim speed:%.2f paused:%v queue:%d", snap.AnimSpeed, snap.AnimPaused, snap.AnimQueue),
	}

	panelH := len(lines)*16 + 8
	r.DrawRect(0, 0, 260, panelH, Color{0, 0, 0, 180})
	for i, line := range lines {
		r.DrawText(line, 11, 4, 4+i*16, Color{0, 255, 0, 255}, false)
	}
}
