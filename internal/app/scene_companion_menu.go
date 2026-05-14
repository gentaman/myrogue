package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
)

type CompanionMenuScene struct {
	game        *GameScene
	companionID entity.ID
	cursor      int
}

var companionOptions = []string{
	"ついてきて (デフォルト)",
	"待ってて (その場に留まる)",
	"積極的に戦って (中立も攻撃)",
	"やめる",
}

func (s *CompanionMenuScene) Update(input InputState) Scene {
	if input.Cancel {
		return s.game
	}

	if input.Up {
		s.cursor--
		if s.cursor < 0 {
			s.cursor = len(companionOptions) - 1
		}
	}
	if input.Down {
		s.cursor++
		if s.cursor >= len(companionOptions) {
			s.cursor = 0
		}
	}

	if input.Confirm {
		aiComp := s.game.GetAI(s.companionID)
		name := s.game.GetName(s.companionID)
		switch s.cursor {
		case 0:
			if aiComp != nil {
				aiComp.Order = component.OrderFollow
			}
			s.game.pushMessage(fmt.Sprintf("%sに「ついてきて」と命令した。", name))
			return s.game
		case 1:
			if aiComp != nil {
				aiComp.Order = component.OrderWait
			}
			s.game.pushMessage(fmt.Sprintf("%sに「待ってて」と命令した。", name))
			return s.game
		case 2:
			if aiComp != nil {
				aiComp.Order = component.OrderAggressive
			}
			s.game.pushMessage(fmt.Sprintf("%sに「積極的に戦って」と命令した。", name))
			return s.game
		case 3:
			return s.game
		}
	}

	return nil
}

func (s *CompanionMenuScene) Draw(r Renderer) {
	s.game.Draw(r)

	name := s.game.GetName(s.companionID)
	aiComp := s.game.GetAI(s.companionID)
	orderName := "ついてきて"
	if aiComp != nil {
		orderName = component.CompanionOrderNames[aiComp.Order]
	}
	title := fmt.Sprintf("%sへの指示 (現在: %s)", name, orderName)

	panelW, panelH := 340, 180
	px, py := (ScreenWidth-panelW)/2, (ScreenHeight-panelH)/2
	r.DrawPanel(px, py, panelW, panelH, Color{30, 30, 40, 255}, Color{200, 200, 200, 255})
	r.DrawText(title, 14, ScreenWidth/2, py+16, Color{255, 255, 255, 255}, true)

	for i, opt := range companionOptions {
		clr := Color{200, 200, 200, 255}
		if i == s.cursor {
			clr = Color{255, 255, 100, 255}
			r.DrawText("▶", 14, px+20, py+52+i*28, clr, false)
		}
		r.DrawText(opt, 14, px+42, py+52+i*28, clr, false)
	}
}
