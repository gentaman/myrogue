package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/action"
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
	playerSkillIDs := game.GetSkills(game.Player)
	for _, id := range playerSkillIDs {
		if sk, ok := game.registry.GetSkillDef(id); ok {
			skills = append(skills, *sk)
		}
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
	pPos := g.playerPos()
	if pPos == nil {
		return g
	}

	var tx, ty int
	if sk.Range <= 1 {
		dx, dy := pPos.Dir.Delta()
		tx, ty = pPos.X+dx, pPos.Y+dy
	} else {
		target, rtx, rty := s.findRangedTarget(sk.Range)
		_ = target
		tx, ty = rtx, rty
	}

	g.Scheduler.IncrementTurn()
	act := &action.SkillAction{
		SkillID: sk.ID,
		TargetX: tx,
		TargetY: ty,
	}
	events := act.Execute(g.Player, g)
	g.processEvents(events)

	g.Scheduler.StartCompanionPhase()
	return g
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
	return entity.InvalidID, pPos.X + dx, pPos.Y + dy
}

func (s *SkillScene) Draw(r Renderer) {
	s.game.Draw(r)
	const (
		w = 300
		h = 240
		x = (ScreenWidth - w) / 2
		y = (ScreenHeight - h) / 2
	)
	r.DrawPanel(x, y, w, h, Color{30, 30, 40, 220}, Color{100, 100, 150, 255})
	r.DrawText("SKILL MENU", 14, ScreenWidth/2, y+15, Color{200, 200, 255, 255}, true)

	for i, sk := range s.skills {
		clr := Color{200, 200, 200, 255}
		if i == s.cursor {
			clr = Color{255, 255, 100, 255}
			r.DrawRect(x+10, y+40+i*20, w-20, 18, Color{255, 255, 255, 30})
		}
		r.DrawText(fmt.Sprintf("%s (MP:%d)", sk.Name, sk.MPCost), 12, x+20, y+40+i*20, clr, false)
	}

	if len(s.skills) > 0 {
		desc := s.skills[s.cursor].Desc
		r.DrawText(desc, 10, ScreenWidth/2, y+h-30, Color{180, 180, 180, 255}, true)
	}
}
