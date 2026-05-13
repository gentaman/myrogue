package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
)

type DebugItemScene struct {
	game   *GameScene
	cursor int
}

func (s *DebugItemScene) Update(input InputState) Scene {
	items := s.game.registry.Items

	if input.Cancel {
		return s.game
	}

	if input.Up {
		s.cursor--
		if s.cursor < 0 {
			s.cursor = len(items) - 1
		}
	}
	if input.Down {
		s.cursor++
		if s.cursor >= len(items) {
			s.cursor = 0
		}
	}

	if input.Confirm {
		itemDef := items[s.cursor]
		inv := s.game.GetInventory(s.game.Player)
		if inv != nil {
			// Find if already exists to stack or just add
			found := false
			for i := range inv.Items {
				if inv.Items[i].DefID == itemDef.ID {
					inv.Items[i].Count++
					found = true
					break
				}
			}
			if !found {
				inv.Items = append(inv.Items, component.ItemEntry{
					DefID:      itemDef.ID,
					Count:      1,
					Durability: itemDef.Durability,
				})
			}
			s.game.pushMessage(fmt.Sprintf("デバッグ: %s を入手しました", itemDef.Name))
		}
		return s.game
	}

	return nil
}

func (s *DebugItemScene) Draw(r Renderer) {
	s.game.Draw(r) // Draw game in background

	items := s.game.registry.Items
	const (
		panelW = 360
		panelH = 400
		rowH   = 24
	)
	panelX, panelY := (ScreenWidth-panelW)/2, (ScreenHeight-panelH)/2
	r.DrawPanel(panelX, panelY, panelW, panelH, Color{20, 20, 30, 240}, Color{100, 100, 200, 255})
	r.DrawText("--- DEBUG: ITEM SPAWNER ---", 14, ScreenWidth/2, panelY+10, Color{255, 255, 0, 255}, true)

	visibleRows := (panelH - 60) / rowH
	scroll := 0
	if s.cursor >= visibleRows {
		scroll = s.cursor - visibleRows + 1
	}

	for i := 0; i < visibleRows; i++ {
		idx := i + scroll
		if idx >= len(items) {
			break
		}
		item := items[idx]
		y := panelY + 40 + i*rowH
		clr := Color{200, 200, 200, 255}
		if idx == s.cursor {
			r.DrawRect(panelX+5, y-2, panelW-10, rowH, Color{60, 60, 100, 255})
			clr = Color{255, 255, 255, 255}
		}
		r.DrawText(fmt.Sprintf("[%s] %s", item.ID, item.Name), 12, panelX+15, y, clr, false)
	}

	r.DrawText("W/S: 選択  Enter: 入手  Esc: 戻る", 12, ScreenWidth/2, panelY+panelH-20, Color{150, 150, 150, 255}, true)
}
