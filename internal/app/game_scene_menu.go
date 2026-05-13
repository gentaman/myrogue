package app

import (
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

func (g *GameScene) buildMenu() {
	g.MenuItems = nil
	pPos := g.playerPos()
	dx, dy := pPos.Dir.Delta()
	fx, fy := pPos.X+dx, pPos.Y+dy

	hasEnemy := false
	occ := g.UnitAt(fx, fy)
	if occ != entity.InvalidID {
		f, _ := g.Factions.Get(occ)
		if f != nil && f.Faction == component.FactionEnemy {
			hasEnemy = true
		}
	}
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "正面を攻撃", Enabled: hasEnemy, Kind: menuAttack})

	hasCompanion := false
	if occ != entity.InvalidID {
		f, _ := g.Factions.Get(occ)
		if f != nil && f.Faction == component.FactionAlly {
			hasCompanion = true
		}
	}
	compLabel := "仲間に指示を出す"
	if hasCompanion {
		compLabel = g.GetName(occ) + "に指示を出す"
	}
	g.MenuItems = append(g.MenuItems, MenuItem{Label: compLabel, Enabled: hasCompanion, Kind: menuCompanion})

	examineLabel := "足元を調べる"
	if g.itemAtFeet() >= 0 {
		examineLabel = "足元のアイテムを調べる"
	}
	g.MenuItems = append(g.MenuItems, MenuItem{Label: examineLabel, Enabled: true, Kind: menuExamine})

	inv := g.GetInventory(g.Player)
	hasItems := inv != nil && len(inv.Items) > 0
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "道具を使う", Enabled: hasItems, Kind: menuItem})
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "セーブ", Enabled: g.SaveService != nil, Kind: menuSave})
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "その場で待機", Enabled: true, Kind: menuWait})
}

func (g *GameScene) execMenuItem() Scene {
	if g.MenuCursor >= len(g.MenuItems) {
		return nil
	}
	item := g.MenuItems[g.MenuCursor]
	if !item.Enabled {
		return nil
	}
	g.MenuOpen = false
	switch item.Kind {
	case menuAttack:
		pPos := g.playerPos()
		dx, dy := pPos.Dir.Delta()
		fx, fy := pPos.X+dx, pPos.Y+dy
		g.Scheduler.IncrementTurn()
		anim, _ := g.Anims.Get(g.Player)
		if anim != nil {
			anim.AttackAnim = component.AttackAnimFrames
		}
		target := g.UnitAt(fx, fy)
		if target != entity.InvalidID {
			g.resolveCombat(g.Player, target)
		}
		g.Scheduler.StartCompanionPhase()
	case menuCompanion:
		pPos := g.playerPos()
		dx, dy := pPos.Dir.Delta()
		fx, fy := pPos.X+dx, pPos.Y+dy
		occ := g.UnitAt(fx, fy)
		if occ != entity.InvalidID {
			return &CompanionMenuScene{game: g, companionID: occ}
		}
	case menuWait:
		g.Scheduler.IncrementTurn()
		g.pushMessage("待機した。")
		g.Scheduler.StartCompanionPhase()
	case menuExamine:
		g.Scheduler.IncrementTurn()
		pPos := g.playerPos()
		tile := g.World.Tiles[pPos.X][pPos.Y]

		if tile == world.Stairs || tile == world.StairsDown || tile == world.StairsUp {
			if tile == world.Stairs {
				g.tryChangeFloor(-1)
			} else if tile == world.StairsDown {
				if g.Audio != nil {
					g.Audio.PlaySFX("stair_down")
				}
				g.tryChangeFloor(1)
			} else if tile == world.StairsUp {
				if g.Audio != nil {
					g.Audio.PlaySFX("stair_up")
				}
				g.tryChangeFloor(-1)
			}
			g.Scheduler.StartCompanionPhase()
			return nil
		}

		if idx := g.itemAtFeet(); idx >= 0 {
			return &ChestScene{game: g, chestIdx: idx}
		}
		g.pushMessage("足元には特に何もない。")
		g.Scheduler.StartCompanionPhase()
	case menuItem:
		return &InventoryScene{game: g}
	case menuSave:
		if g.SaveService != nil {
			if err := g.SaveService.Save("slot1", g); err != nil {
				g.pushMessage("セーブに失敗した...")
			} else {
				g.pushMessage("セーブ完了！")
			}
		}
	}
	return nil
}

func (g *GameScene) itemAtFeet() int {
	pPos := g.playerPos()
	for i, it := range g.World.Items {
		if it.X == pPos.X && it.Y == pPos.Y {
			return i
		}
	}
	return -1
}
