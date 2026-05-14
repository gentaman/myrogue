package app

import (
	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
)

func (g *GameScene) useInventoryItem(idx int) bool {
	act := &action.UseItemAction{ItemIdx: idx}
	events := act.Execute(g.Player, g)
	if len(events) == 0 {
		return false
	}
	g.processEvents(events)
	return true
}

func (g *GameScene) dropInventoryItem(idx int) {
	act := &action.DropItemAction{ItemIdx: idx}
	events := act.Execute(g.Player, g)
	g.processEvents(events)
}

func (g *GameScene) isItemPossessed(itemID string) bool {
	// プレイヤーのインベントリ
	inv, ok := g.Inventories.Get(g.Player)
	if ok {
		for _, it := range inv.Items {
			if it.DefID == itemID {
				return true
			}
		}
	}
	// 仲間のインベントリ
	allies := g.EntitiesWithFaction(component.FactionAlly)
	for _, id := range allies {
		inv, ok := g.Inventories.Get(id)
		if ok {
			for _, it := range inv.Items {
				if it.DefID == itemID {
					return true
				}
			}
		}
	}
	return false
}
