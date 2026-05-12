package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
)

type invSubMenu int

const (
	invSubNone invSubMenu = iota
	invSubUse
	invSubDrop
	invSubDesc
)

type InventoryScene struct {
	game         *GameScene
	targetID     int
	cursor       int
	subMenu      invSubMenu
	subCursor    int
	scrollOffset int
}

var subMenuLabels = []string{"使う", "捨てる", "説明を見る", "キャンセル"}

func (s *InventoryScene) getInventory() *component.Inventory {
	return s.game.GetInventory(s.game.Player)
}

func (s *InventoryScene) clampCursor() {
	inv := s.getInventory()
	if inv == nil {
		s.cursor = 0
		return
	}
	n := len(inv.Items)
	if n == 0 {
		s.cursor = 0
		return
	}
	if s.cursor >= n {
		s.cursor = n - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
}

func (s *InventoryScene) syncScroll(visibleRows int) {
	if s.cursor < s.scrollOffset {
		s.scrollOffset = s.cursor
	}
	if s.cursor >= s.scrollOffset+visibleRows {
		s.scrollOffset = s.cursor - visibleRows + 1
	}
}

func (s *InventoryScene) Update(input InputState) Scene {
	inv := s.getInventory()
	if inv == nil {
		return s.game
	}

	if s.subMenu == invSubDesc {
		if input.Cancel || input.Confirm {
			s.subMenu = invSubNone
		}
		return nil
	}

	if s.subMenu != invSubNone {
		if input.Cancel {
			s.subMenu = invSubNone
			return nil
		}
		if input.Up {
			s.subCursor = (s.subCursor - 1 + len(subMenuLabels)) % len(subMenuLabels)
		}
		if input.Down {
			s.subCursor = (s.subCursor + 1) % len(subMenuLabels)
		}
		if input.Confirm {
			switch s.subCursor {
			case 0: // 使う
				if s.game.useInventoryItem(s.cursor) {
					s.clampCursor()
					s.subMenu = invSubNone
					s.game.Scheduler.IncrementTurn()
					s.game.Scheduler.StartCompanionPhase()
					return s.game
				}
				s.subMenu = invSubNone
			case 1: // 捨てる
				s.game.dropInventoryItem(s.cursor)
				s.clampCursor()
				s.subMenu = invSubNone
			case 2: // 説明を見る
				s.subMenu = invSubDesc
			case 3: // キャンセル
				s.subMenu = invSubNone
			}
		}
		return nil
	}

	if input.Cancel || input.Inventory {
		return s.game
	}

	if len(inv.Items) == 0 {
		return nil
	}

	visibleRows := s.calcVisibleRows()
	if input.Up {
		s.cursor = (s.cursor - 1 + len(inv.Items)) % len(inv.Items)
		s.syncScroll(visibleRows)
	}
	if input.Down {
		s.cursor = (s.cursor + 1) % len(inv.Items)
		s.syncScroll(visibleRows)
	}
	if input.Confirm {
		s.subMenu = invSubUse
		s.subCursor = 0
	}
	return nil
}

func (s *InventoryScene) calcVisibleRows() int {
	const (
		rowH    = 32
		headerH = 44
		footerH = 28
	)
	panelH := headerH + 10*rowH + footerH
	if panelH > ScreenHeight-40 {
		panelH = ScreenHeight - 40
	}
	visibleRows := (panelH - headerH - footerH) / rowH
	if visibleRows < 1 {
		visibleRows = 1
	}
	return visibleRows
}

func (s *InventoryScene) Draw(r Renderer) {
	s.game.Draw(r)

	const (
		rowH     = 32
		headerH  = 44
		footerH  = 28
		padX     = 58
		padRight = 16
		minW     = 300
	)

	inv := s.getInventory()
	if inv == nil {
		return
	}

	panelW := minW
	for _, entry := range inv.Items {
		def, ok := s.game.registry.GetItemDef(entry.DefID)
		if !ok {
			continue
		}
		label := fmt.Sprintf("%s  x%d  (重%d)", def.Name, entry.Count, def.Weight)
		w := r.MeasureText(label, 14) + padX + padRight
		if w > panelW {
			panelW = w
		}
	}
	if panelW > ScreenWidth-20 {
		panelW = ScreenWidth - 20
	}

	n := len(inv.Items)
	if n == 0 {
		n = 1
	}
	panelH := headerH + n*rowH + footerH
	if panelH > ScreenHeight-40 {
		panelH = ScreenHeight - 40
	}
	visibleRows := (panelH - headerH - footerH) / rowH
	if visibleRows < 1 {
		visibleRows = 1
	}

	panelX := (ScreenWidth - panelW) / 2
	panelY := (ScreenHeight - panelH) / 2
	r.DrawPanel(panelX, panelY, panelW, panelH, Color{15, 20, 40, 255}, Color{80, 180, 100, 255})
	r.DrawText("アイテム", 14, ScreenWidth/2, panelY+12, Color{150, 255, 150, 255}, true)

	if len(inv.Items) == 0 {
		r.DrawText("アイテムを持っていない", 14, ScreenWidth/2, panelY+panelH/2-10, Color{120, 120, 120, 255}, true)
	} else {
		for i, entry := range inv.Items {
			row := i - s.scrollOffset
			if row < 0 || row >= visibleRows {
				continue
			}
			y := panelY + headerH + row*rowH
			def, ok := s.game.registry.GetItemDef(entry.DefID)
			clr := Color{200, 200, 200, 255}
			if i == s.cursor {
				r.DrawRect(panelX+4, y-4, panelW-8, 26, Color{40, 80, 50, 255})
				clr = Color{255, 255, 255, 255}
				r.DrawText("▶", 12, panelX+14, y, clr, false)
			}
			label := "???"
			if ok {
				eqMark := ""
				if entry.Equipped {
					eqMark = "[E] "
				}
				label = fmt.Sprintf("%s%s  x%d  (重%d)", eqMark, def.Name, entry.Count, def.Weight)
			}
			r.DrawText(label, 14, panelX+44, y, clr, false)
		}
	}

	if s.subMenu == invSubDesc {
		s.drawDescWindow(r, panelX, panelY, panelW, panelH)
	} else if s.subMenu != invSubNone {
		s.drawSubMenu(r, panelX, panelY, panelW)
	}

	hint := "W/S: 選択  Enter: アクション  I/Esc: 閉じる"
	if s.subMenu == invSubDesc {
		hint = "Enter/Esc: 閉じる"
	} else if s.subMenu != invSubNone {
		hint = "W/S: 選択  Enter: 決定  Esc: キャンセル"
	}
	r.DrawText(hint, 12, ScreenWidth/2, panelY+panelH-20, Color{100, 120, 100, 255}, true)
}

func (s *InventoryScene) drawSubMenu(r Renderer, panelX, panelY, panelW int) {
	const (
		subRowH   = 26
		subPadX   = 36
		subPadR   = 16
		subPadTop = 14
		subPadBot = 10
		minSubW   = 120
	)
	subW := minSubW
	for _, label := range subMenuLabels {
		w := r.MeasureText(label, 14) + subPadX + subPadR
		if w > subW {
			subW = w
		}
	}
	subH := subPadTop + len(subMenuLabels)*subRowH + subPadBot
	subX := panelX + (panelW-subW)/2
	row := s.cursor - s.scrollOffset
	if row < 0 {
		row = 0
	}
	subY := panelY + 44 + row*32 - 8
	if subY+subH > ScreenHeight-4 {
		subY = ScreenHeight - subH - 4
	}
	if subY < 4 {
		subY = 4
	}
	r.DrawPanel(subX, subY, subW, subH, Color{30, 30, 60, 255}, Color{100, 150, 255, 255})
	for i, label := range subMenuLabels {
		y := subY + subPadTop + i*subRowH
		clr := Color{200, 200, 200, 255}
		if i == s.subCursor {
			r.DrawRect(subX+4, y-3, subW-8, 22, Color{60, 60, 120, 255})
			clr = Color{255, 255, 255, 255}
			r.DrawText("▶", 12, subX+10, y, clr, false)
		}
		r.DrawText(label, 14, subX+26, y, clr, false)
	}
}

func (s *InventoryScene) drawDescWindow(r Renderer, panelX, panelY, panelW, panelH int) {
	inv := s.getInventory()
	if inv == nil || s.cursor < 0 || s.cursor >= len(inv.Items) {
		return
	}
	entry := inv.Items[s.cursor]
	def, ok := s.game.registry.GetItemDef(entry.DefID)
	if !ok {
		return
	}
	const (
		dPad  = 16
		lineH = 20
	)
	dW := panelW - dPad*2
	if dW < 200 {
		dW = 200
	}
	dH := 28 + lineH*3 + dPad*2
	dX, dY := panelX+dPad, panelY+(panelH-dH)/2
	if dY < panelY+4 {
		dY = panelY + 4
	}
	r.DrawPanel(dX, dY, dW, dH, Color{20, 20, 50, 255}, Color{150, 200, 255, 255})
	r.DrawText(def.Name, 14, dX+dW/2, dY+8, Color{200, 230, 255, 255}, true)
	statsY := dY + 28
	r.DrawText(fmt.Sprintf("重量: %d", def.Weight), 12, dX+dPad, statsY, Color{160, 220, 160, 255}, false)
	if def.Desc != "" {
		r.DrawText(def.Desc, 12, dX+dPad, statsY+lineH, Color{200, 200, 200, 255}, false)
	}
}
