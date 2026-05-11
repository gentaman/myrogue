package app

import (
	"fmt"
	"math/rand"

	"github.com/gentaman/myrogue/internal/animation"
	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/combat"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

type SkillScene struct {
	game   *GameScene
	cursor int
	skills []content.SkillDef
}

func NewSkillScene(game *GameScene) *SkillScene {
	var skills []content.SkillDef
	for _, sk := range game.Registry.Skills {
		skills = append(skills, sk)
	}
	return &SkillScene{game: game, skills: skills}
}

func (s *SkillScene) Update(input InputState) Scene {
	if input.Cancel {
		return s.game
	}
	if input.Up {
		s.cursor--
		if s.cursor < 0 {
			s.cursor = len(s.skills) - 1
		}
	}
	if input.Down {
		s.cursor++
		if s.cursor >= len(s.skills) {
			s.cursor = 0
		}
	}
	if input.Confirm && len(s.skills) > 0 {
		sk := s.skills[s.cursor]
		return s.useSkill(sk)
	}
	return nil
}

func (s *SkillScene) useSkill(sk content.SkillDef) Scene {
	g := s.game
	stats := g.GetStats(g.Player)
	if stats == nil {
		return g
	}
	if stats.MP < sk.MPCost {
		g.pushMessage("MPが足りない！")
		return g
	}
	stats.MP -= sk.MPCost
	g.Scheduler.IncrementTurn()

	switch sk.Type {
	case content.SkillTypeAttack:
		s.executeAttackSkill(sk)
	case content.SkillTypeHeal:
		heal := sk.Power + stats.Wis/2
		stats.HP += heal
		if stats.HP > stats.MaxHP {
			stats.HP = stats.MaxHP
		}
		g.pushMessage(fmt.Sprintf("%sを使った！ HPが%d回復した。", sk.Name, heal))
	case content.SkillTypeBuff:
		if sk.Status != "" {
			g.applyStatus(g.Player, component.StatusFromString(sk.Status), sk.Duration)
		}
		g.pushMessage(fmt.Sprintf("%sを使った！", sk.Name))
	case content.SkillTypeDebuff:
		s.executeDebuffSkill(sk)
	}

	g.Scheduler.StartCompanionPhase()
	return g
}

func (s *SkillScene) executeAttackSkill(sk content.SkillDef) {
	g := s.game
	pPos := g.playerPos()
	if pPos == nil {
		return
	}

	var target entity.ID
	var tx, ty int

	if sk.Range <= 1 {
		dx, dy := pPos.Dir.Delta()
		tx, ty = pPos.X+dx, pPos.Y+dy
		target = g.UnitAt(tx, ty)
	} else {
		target, tx, ty = s.findRangedTarget(sk.Range)
	}

	if target == entity.InvalidID {
		g.pushMessage(fmt.Sprintf("%sを放った！ しかし何もいなかった。", sk.Name))
		return
	}

	if sk.Range > 1 {
		g.AnimQueue.Add(animation.Projectile{
			StartX:      float64(pPos.X * TileSize),
			StartY:      float64(pPos.Y * TileSize),
			EndX:        float64(tx * TileSize),
			EndY:        float64(ty * TileSize),
			TotalFrames: 12,
			ColorHex:    sk.ColorHex,
		})
	}

	attacker := g.buildCombatant(g.Player)
	defender := g.buildCombatant(target)
	damage := combat.CalcDamage(attacker, defender, combat.CombatTypeMagical, sk.Power, sk.Element)
	if damage < 1 {
		damage = 1
	}

	defStats := g.GetStats(target)
	if defStats != nil {
		defStats.HP -= damage
	}

	g.pushMessage(fmt.Sprintf("%sで%sに %d ダメージ！", sk.Name, g.GetName(target), damage))

	if defStats != nil && defStats.HP <= 0 {
		events := []action.Event{action.EventDeath{Entity: target, Name: g.GetName(target)}}
		g.processEvents(events)
	}
}

func (s *SkillScene) executeDebuffSkill(sk content.SkillDef) {
	g := s.game
	pPos := g.playerPos()
	if pPos == nil {
		return
	}

	var target entity.ID
	var tx, ty int

	if sk.Range <= 1 {
		dx, dy := pPos.Dir.Delta()
		tx, ty = pPos.X+dx, pPos.Y+dy
		target = g.UnitAt(tx, ty)
	} else {
		target, tx, ty = s.findRangedTarget(sk.Range)
	}

	if target == entity.InvalidID {
		g.pushMessage(fmt.Sprintf("%sを放った！ しかし何もいなかった。", sk.Name))
		return
	}

	if sk.Range > 1 {
		g.AnimQueue.Add(animation.Projectile{
			StartX:      float64(pPos.X * TileSize),
			StartY:      float64(pPos.Y * TileSize),
			EndX:        float64(tx * TileSize),
			EndY:        float64(ty * TileSize),
			TotalFrames: 12,
			ColorHex:    sk.ColorHex,
		})
	}

	if sk.Power > 0 {
		attacker := g.buildCombatant(g.Player)
		defender := g.buildCombatant(target)
		damage := combat.CalcDamage(attacker, defender, combat.CombatTypeMagical, sk.Power, sk.Element)
		defStats := g.GetStats(target)
		if defStats != nil {
			defStats.HP -= damage
			g.pushMessage(fmt.Sprintf("%sで%sに %d ダメージ！", sk.Name, g.GetName(target), damage))
		}
	}

	if sk.Status != "" && rand.Intn(100) < 70 {
		g.applyStatus(target, component.StatusFromString(sk.Status), sk.Duration)
	} else if sk.Status != "" {
		g.pushMessage(fmt.Sprintf("%sは効かなかった...", sk.Name))
	}
}

func (s *SkillScene) findRangedTarget(maxRange int) (entity.ID, int, int) {
	g := s.game
	pPos := g.playerPos()
	dx, dy := pPos.Dir.Delta()

	for dist := 1; dist <= maxRange; dist++ {
		tx, ty := pPos.X+dx*dist, pPos.Y+dy*dist
		if !g.World.InBounds(tx, ty) || g.World.Tiles[tx][ty] == world.Wall {
			break
		}
		occ := g.UnitAt(tx, ty)
		if occ != entity.InvalidID {
			return occ, tx, ty
		}
	}
	return entity.InvalidID, 0, 0
}

func (s *SkillScene) Draw(r Renderer) {
	s.game.Draw(r)

	panelW, panelH := 300, 40+len(s.skills)*26
	if panelH > ScreenHeight-40 {
		panelH = ScreenHeight - 40
	}
	px, py := (ScreenWidth-panelW)/2, 30
	r.DrawPanel(px, py, panelW, panelH, Color{20, 20, 40, 240}, Color{100, 100, 200, 255})
	r.DrawText("スキル選択", 14, ScreenWidth/2, py+12, Color{255, 220, 100, 255}, true)

	stats := s.game.GetStats(s.game.Player)
	mpText := ""
	if stats != nil {
		mpText = fmt.Sprintf("MP: %d/%d", stats.MP, stats.MaxMP)
	}
	r.DrawText(mpText, 11, px+panelW-10, py+12, Color{100, 200, 255, 255}, false)

	for i, sk := range s.skills {
		y := py + 36 + i*26
		clr := Color{200, 200, 200, 255}
		if stats != nil && stats.MP < sk.MPCost {
			clr = Color{100, 100, 100, 255}
		}
		if i == s.cursor {
			clr = Color{255, 255, 100, 255}
			r.DrawText("▶", 14, px+10, y, clr, false)
		}
		label := fmt.Sprintf("%s (MP:%d)", sk.Name, sk.MPCost)
		r.DrawText(label, 13, px+30, y, clr, false)
	}
}
