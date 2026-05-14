package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
)

type ChestScene struct {
	game        *GameScene
	chestIdx    int
	focusLeft   bool
	leftCursor  int
	rightCursor int
}

func (s *ChestScene) Update(input InputState) Scene {
	if input.Cancel {
		return s.game
	}

	chest := &s.game.World.Items[s.chestIdx]
	inv := s.game.GetInventory(s.game.Player)
	if inv == nil {
		return s.game
	}

	if input.Left || input.Right {
		s.focusLeft = !s.focusLeft
	}

	if s.focusLeft {
		n := len(chest.Inventory)
		if n > 0 {
			if input.Up {
				s.leftCursor = (s.leftCursor - 1 + n) % n
			}
			if input.Down {
				s.leftCursor = (s.leftCursor + 1) % n
			}
		}
	} else {
		n := len(inv.Items)
		if n > 0 {
			if input.Up {
				s.rightCursor = (s.rightCursor - 1 + n) % n
			}
			if input.Down {
				s.rightCursor = (s.rightCursor + 1) % n
			}
		}
	}

	if input.Confirm {
		if s.focusLeft {
			// Chest -> Player
			if len(chest.Inventory) > 0 && s.leftCursor < len(chest.Inventory) {
				entry := chest.Inventory[s.leftCursor]
				def, ok := s.game.registry.GetItemDef(entry.DefID)
				weight := 0
				if ok {
					weight = def.Weight
				}
				currentWeight := s.game.currentWeight()
				if inv.MaxCarryWeight > 0 && currentWeight+weight > inv.MaxCarryWeight {
					s.game.pushMessage("重すぎて持てない！")
				} else {
					inv.Items = append(inv.Items, entry)
					chest.Inventory = append(chest.Inventory[:s.leftCursor], chest.Inventory[s.leftCursor+1:]...)
					if s.leftCursor >= len(chest.Inventory) && s.leftCursor > 0 {
						s.leftCursor--
					}
					if ok {
						s.game.pushMessage(def.Name + "を拾った。")
					}
				}
			}
		} else {
			// Player -> Chest
			if len(inv.Items) > 0 && s.rightCursor < len(inv.Items) {
				entry := inv.Items[s.rightCursor]
				if entry.Equipped {
					s.game.pushMessage("装備中は入れられない。")
				} else {
					chest.Inventory = append(chest.Inventory, entry)
					inv.Items = append(inv.Items[:s.rightCursor], inv.Items[s.rightCursor+1:]...)
					if s.rightCursor >= len(inv.Items) && s.rightCursor > 0 {
						s.rightCursor--
					}
				}
			}
		}
	}

	// Remove empty chest
	if len(chest.Inventory) == 0 {
		s.game.World.Items = append(s.game.World.Items[:s.chestIdx], s.game.World.Items[s.chestIdx+1:]...)
		return s.game
	}

	return nil
}

func (s *ChestScene) Draw(r Renderer) {
	s.game.Draw(r)

	const (
		panelW  = 260
		panelH  = 300
		gap     = 16
		rowH    = 24
		headerH = 40
	)
	panelY := (ScreenHeight - panelH) / 2
	leftX := (ScreenWidth - panelW*2 - gap) / 2
	rightX := leftX + panelW + gap

	chest := &s.game.World.Items[s.chestIdx]
	inv := s.game.GetInventory(s.game.Player)

	// Left panel: chest
	lBorder := Color{60, 60, 100, 255}
	if s.focusLeft {
		lBorder = Color{100, 100, 200, 255}
	}
	r.DrawPanel(leftX, panelY, panelW, panelH, Color{20, 20, 30, 255}, lBorder)
	r.DrawText("宝箱の中身", 14, leftX+panelW/2, panelY+12, Color{255, 220, 100, 255}, true)

	for i, entry := range chest.Inventory {
		y := panelY + headerH + i*rowH
		clr := Color{180, 180, 180, 255}
		if s.focusLeft && i == s.leftCursor {
			r.DrawText("▶", 12, leftX+10, y, Color{255, 255, 255, 255}, false)
			clr = Color{255, 255, 255, 255}
		}
		name := entry.DefID
		def, ok := s.game.registry.GetItemDef(entry.DefID)
		if ok {
			name = def.Name
			if def.Durability > 0 {
				name = fmt.Sprintf("%s [%d/%d]", def.Name, entry.Durability, def.Durability)
			}
		}
		r.DrawText(name, 12, leftX+30, y, clr, false)
	}

	// Right panel: player inventory
	rBorder := Color{60, 60, 100, 255}
	if !s.focusLeft {
		rBorder = Color{100, 100, 200, 255}
	}
	r.DrawPanel(rightX, panelY, panelW, panelH, Color{20, 20, 30, 255}, rBorder)
	weightStr := fmt.Sprintf("持物 (%d/%d)", s.game.currentWeight(), inv.MaxCarryWeight)
	r.DrawText(weightStr, 14, rightX+panelW/2, panelY+12, Color{150, 255, 150, 255}, true)

	if inv != nil {
		for i, entry := range inv.Items {
			y := panelY + headerH + i*rowH
			clr := Color{180, 180, 180, 255}
			if !s.focusLeft && i == s.rightCursor {
				r.DrawText("▶", 12, rightX+10, y, Color{255, 255, 255, 255}, false)
				clr = Color{255, 255, 255, 255}
			}
			name := entry.DefID
			def, ok := s.game.registry.GetItemDef(entry.DefID)
			if ok {
				name = def.Name
				if def.Durability > 0 {
					name = fmt.Sprintf("%s [%d/%d]", def.Name, entry.Durability, def.Durability)
				}
			}
			if entry.Equipped {
				name = "[E] " + name
			}
			r.DrawText(name, 12, rightX+30, y, clr, false)
		}
	}

	r.DrawText("←→: 切替  Enter: 移動  Esc: 閉じる", 12, ScreenWidth/2, panelY+panelH+12, Color{180, 180, 180, 255}, true)
}

func (g *GameScene) currentWeight() int {
	inv := g.GetInventory(g.Player)
	if inv == nil {
		return 0
	}
	w := 0
	for _, entry := range inv.Items {
		def, ok := g.registry.GetItemDef(entry.DefID)
		if ok {
			w += def.Weight * entry.Count
		}
	}
	return w
}

func (g *GameScene) addItemToInventory(entry component.ItemEntry) bool {
	inv := g.GetInventory(g.Player)
	if inv == nil {
		return false
	}
	def, ok := g.registry.GetItemDef(entry.DefID)
	if !ok {
		return false
	}
	if inv.MaxCarryWeight > 0 && g.currentWeight()+def.Weight > inv.MaxCarryWeight {
		return false
	}
	inv.Items = append(inv.Items, entry)
	return true
}
