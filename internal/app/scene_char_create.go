package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/save"
)

type CharCreateScene struct {
	registry    *content.Registry
	audio       AudioPlayer
	saveService *save.Service
	phase       int // 0=race, 1=element, 2=confirm
	raceCursor  int
	elemCursor  int
}

var selectableRaces = []component.Race{
	component.RaceHuman,
	component.RaceElf,
	component.RaceDwarf,
	component.RaceGnome,
	component.RaceHalfling,
}

var raceBonus = map[component.Race][6]int{
	component.RaceHuman:    {0, 0, 0, 0, 0, 2},
	component.RaceElf:      {-1, 2, 1, -1, 2, 0},
	component.RaceDwarf:    {2, -1, 0, 3, -2, 0},
	component.RaceGnome:    {-1, 3, 2, -1, 0, 0},
	component.RaceHalfling: {-1, 0, 0, -1, 3, 2},
}

var selectableElements = []component.Element{
	component.ElementNone,
	component.ElementFire,
	component.ElementWater,
	component.ElementAir,
	component.ElementEarth,
	component.ElementLight,
	component.ElementDark,
}

var elementNames = map[component.Element]string{
	component.ElementNone:  "無属性",
	component.ElementFire:  "火",
	component.ElementWater: "水",
	component.ElementAir:   "風",
	component.ElementEarth: "地",
	component.ElementLight: "光",
	component.ElementDark:  "闇",
}

func NewCharCreateScene(reg *content.Registry, audio AudioPlayer, ss *save.Service) *CharCreateScene {
	return &CharCreateScene{registry: reg, audio: audio, saveService: ss}
}

func (s *CharCreateScene) Update(input InputState) Scene {
	switch s.phase {
	case 0:
		if input.Cancel {
			return NewTitleSceneWithDeps(s.registry, s.audio, s.saveService)
		}
		if input.Up {
			s.raceCursor--
			if s.raceCursor < 0 {
				s.raceCursor = len(selectableRaces) - 1
			}
		}
		if input.Down {
			s.raceCursor++
			if s.raceCursor >= len(selectableRaces) {
				s.raceCursor = 0
			}
		}
		if input.Confirm {
			s.phase = 1
		}
	case 1:
		if input.Cancel {
			s.phase = 0
		}
		if input.Up {
			s.elemCursor--
			if s.elemCursor < 0 {
				s.elemCursor = len(selectableElements) - 1
			}
		}
		if input.Down {
			s.elemCursor++
			if s.elemCursor >= len(selectableElements) {
				s.elemCursor = 0
			}
		}
		if input.Confirm {
			s.phase = 2
		}
	case 2:
		if input.Cancel {
			s.phase = 1
		}
		if input.Confirm {
			return s.startGame()
		}
	}
	return nil
}

func (s *CharCreateScene) startGame() Scene {
	race := selectableRaces[s.raceCursor]
	elem := selectableElements[s.elemCursor]

	InitializePlayer(s.registry, race, elem)

	return NewGameScene(s.registry, s.audio, s.saveService)
}

func (s *CharCreateScene) Draw(r Renderer) {
	r.Clear(10, 10, 30, 255)
	r.DrawText("キャラクター作成", 20, ScreenWidth/2, 30, Color{255, 220, 100, 255}, true)

	switch s.phase {
	case 0:
		r.DrawText("種族を選択", 14, ScreenWidth/2, 70, Color{200, 200, 200, 255}, true)
		for i, race := range selectableRaces {
			y := 110 + i*32
			clr := Color{200, 200, 200, 255}
			if i == s.raceCursor {
				clr = Color{255, 255, 100, 255}
				r.DrawText("▶", 14, 140, y, clr, false)
			}
			bonus := raceBonus[race]
			label := fmt.Sprintf("%s  STR%+d WIS%+d FAI%+d VIT%+d AGI%+d LUK%+d",
				component.RaceNames[race], bonus[0], bonus[1], bonus[2], bonus[3], bonus[4], bonus[5])
			r.DrawText(label, 12, 160, y, clr, false)
		}
		r.DrawText("Enter:決定  Esc:戻る", 11, ScreenWidth/2, ScreenHeight-30, Color{120, 120, 120, 255}, true)

	case 1:
		r.DrawText("属性を選択", 14, ScreenWidth/2, 70, Color{200, 200, 200, 255}, true)
		for i, elem := range selectableElements {
			y := 110 + i*28
			clr := Color{200, 200, 200, 255}
			if i == s.elemCursor {
				clr = Color{255, 255, 100, 255}
				r.DrawText("▶", 14, 200, y, clr, false)
			}
			r.DrawText(elementNames[elem], 14, 220, y, clr, false)
		}
		r.DrawText("Enter:決定  Esc:戻る", 11, ScreenWidth/2, ScreenHeight-30, Color{120, 120, 120, 255}, true)

	case 2:
		race := selectableRaces[s.raceCursor]
		elem := selectableElements[s.elemCursor]
		r.DrawText("確認", 14, ScreenWidth/2, 70, Color{200, 200, 200, 255}, true)
		r.DrawText(fmt.Sprintf("種族: %s", component.RaceNames[race]), 14, ScreenWidth/2, 130, Color{255, 255, 255, 255}, true)
		r.DrawText(fmt.Sprintf("属性: %s", elementNames[elem]), 14, ScreenWidth/2, 160, Color{255, 255, 255, 255}, true)

		bonus := raceBonus[race]
		p := &s.registry.Players[0]
		base := [6]int{p.Str, p.Wis, p.Fai, p.Vit, p.Agi, p.Luk}
		r.DrawText(fmt.Sprintf("STR:%d WIS:%d FAI:%d VIT:%d AGI:%d LUK:%d",
			base[0]+bonus[0], base[1]+bonus[1], base[2]+bonus[2],
			base[3]+bonus[3], base[4]+bonus[4], base[5]+bonus[5]),
			12, ScreenWidth/2, 200, Color{180, 220, 255, 255}, true)

		r.DrawText("Enter:ゲーム開始  Esc:戻る", 11, ScreenWidth/2, ScreenHeight-30, Color{120, 120, 120, 255}, true)
	}
}
