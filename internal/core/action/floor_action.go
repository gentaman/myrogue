package action

import (
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
)

type FloorChangeAction struct {
	Direction int
}

func (a *FloorChangeAction) Execute(actor entity.ID, w WorldAccess) []event.Event {
	currentFloor := w.Map().Floor
	if currentFloor == 0 && a.Direction < 0 {
		// Attempting to exit the dungeon from the first floor
		inv := w.GetInventory(actor)
		hasTreasure := false
		if inv != nil {
			reg := w.Registry()
			for _, entry := range inv.Items {
				if def, ok := reg.GetItemDef(entry.DefID); ok && def.FloorBound {
					hasTreasure = true
					break
				}
			}
		}

		if hasTreasure {
			return []event.Event{event.EventWin{}}
		} else {
			return []event.Event{event.EventLog{Text: "宝がない！まだ帰れない。"}}
		}
	}

	return []event.Event{event.EventFloorChange{
		CurrentFloor: currentFloor,
		Direction:    a.Direction,
	}}
}
