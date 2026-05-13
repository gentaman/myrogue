package action

import (
	"fmt"
	"math"
	"strings"

	"github.com/gentaman/myrogue/internal/core/combat"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
	"github.com/gentaman/myrogue/internal/core/world"
)

type UseItemAction struct {
	ItemIdx int
}

func (a *UseItemAction) Execute(actor entity.ID, w WorldAccess) []event.Event {
	inv := w.GetInventory(actor)
	reg := w.Registry()
	if inv == nil || reg == nil || a.ItemIdx < 0 || a.ItemIdx >= len(inv.Items) {
		return nil
	}
	entry := inv.Items[a.ItemIdx]
	def, ok := reg.GetItemDef(entry.DefID)
	if !ok {
		return nil
	}

	var events []event.Event

	if def.EquipSlot != component.SlotNone {
		events = append(events, event.EventEquip{Entity: actor, ItemIdx: a.ItemIdx})
		if !entry.Equipped {
			events = append(events, event.EventLog{Text: def.Name + "を装備した。"})
		} else {
			events = append(events, event.EventLog{Text: def.Name + "を外した。"})
		}
		return events
	}

	if def.Effect != nil {
		switch def.Effect.Type {
		case "heal":
			events = append(events, event.EventHeal{Entity: actor, Amount: def.Effect.Amount})
			events = append(events, event.EventLog{Text: fmt.Sprintf("%sを使った。HPが%d回復した。", def.Name, def.Effect.Amount)})
		case "mp_heal":
			events = append(events, event.EventMP{Entity: actor, Amount: def.Effect.Amount})
			events = append(events, event.EventLog{Text: fmt.Sprintf("%sを使った。MPが%d回復した。", def.Name, def.Effect.Amount)})
		case "reveal_map":
			events = append(events, a.handleRevealMap(entry, def, w)...)
		case "fireball", "ranged_magic":
			events = append(events, a.handleRangedMagic(actor, def, w)...)
		case "area_magic":
			events = append(events, a.handleAreaMagic(actor, def, w)...)
		case "heal_faith":
			events = append(events, a.handleHealFaith(actor, def, w)...)
		default:
			events = append(events, event.EventLog{Text: def.Name + "を使った。"})
		}
	} else {
		events = append(events, event.EventLog{Text: def.Name + "を使った。"})
	}

	if len(events) > 0 {
		if def.Durability > 1 {
			events = append(events, event.EventItemConsume{Entity: actor, ItemIdx: a.ItemIdx, Durability: 1})
		} else {
			events = append(events, event.EventItemConsume{Entity: actor, ItemIdx: a.ItemIdx, Count: 1})
		}
	}
	return events
}

func (a *UseItemAction) handleRevealMap(entry component.ItemEntry, def *content.ItemDef, w WorldAccess) []event.Event {
	if def.Condition != nil && def.Condition.Type == "check_same_floor" {
		if entry.ObtainedFloor != w.Map().Floor {
			return []event.Event{event.EventLog{Text: fmt.Sprintf(def.Condition.Message, entry.ObtainedFloor+1)}}
		}
	}
	var events []event.Event
	events = append(events, event.EventRevealMap{})
	msg := def.Effect.Message
	if strings.Contains(msg, "%d") {
		msg = fmt.Sprintf(msg, w.Map().Floor+1)
	}
	events = append(events, event.EventLog{Text: msg})
	return events
}

func (a *UseItemAction) handleRangedMagic(actor entity.ID, def *content.ItemDef, w WorldAccess) []event.Event {
	stats := w.GetStats(actor)
	if stats == nil || stats.MP < def.Effect.Cost {
		return []event.Event{event.EventLog{Text: "MPが足りない！"}}
	}

	var events []event.Event
	if def.Effect.Cost > 0 {
		events = append(events, event.EventMP{Entity: actor, Amount: -def.Effect.Cost})
	}

	pos := w.GetPosition(actor)
	dx, dy := pos.Dir.Delta()
	rangeLimit := def.Effect.Range
	if rangeLimit == 0 {
		rangeLimit = 1
	}

	var targetID entity.ID = entity.InvalidID
	tx, ty := pos.X, pos.Y
	for i := 1; i <= rangeLimit; i++ {
		currX, currY := pos.X+dx*i, pos.Y+dy*i
		if !w.Map().InBounds(currX, currY) || w.Map().Tiles[currX][currY] == world.Wall {
			break
		}
		tx, ty = currX, currY
		occ := w.UnitAt(currX, currY)
		if occ != entity.InvalidID {
			targetID = occ
			break
		}
	}

	color := def.Effect.Color
	if color == "" {
		color = "#FF4500"
	}

	events = append(events, event.EventProjectile{
		StartX: float64(pos.X), StartY: float64(pos.Y),
		EndX: float64(tx), EndY: float64(ty),
		TotalFrames: 15,
		ColorHex:    color,
	})

	if targetID != entity.InvalidID {
		atk := BuildCombatant(actor, w)
		defCombatant := BuildCombatant(targetID, w)
		res := combat.ResolveCombat(atk, defCombatant, combat.CombatTypeMagical, def.Effect.Damage, component.ElementFromString(def.Effect.Element), w.PlayerID(), w.RNG())

		for _, log := range res.Log {
			events = append(events, event.EventLog{Text: log})
		}
		events = append(events, event.EventAttack{
			Attacker: actor, Defender: targetID,
			Damage: res.Damage, Missed: res.Missed, Dir: pos.Dir,
		})
		if res.Killed {
			events = append(events, event.EventDeath{Entity: targetID, Name: defCombatant.Name})
			if res.XP > 0 {
				events = append(events, event.EventXP{Entity: actor, Amount: res.XP})
			}
		}
	} else {
		events = append(events, event.EventLog{Text: def.Effect.Message})
	}
	return events
}

func (a *UseItemAction) handleAreaMagic(actor entity.ID, def *content.ItemDef, w WorldAccess) []event.Event {
	stats := w.GetStats(actor)
	if stats == nil || stats.MP < def.Effect.Cost {
		return []event.Event{event.EventLog{Text: "MPが足りない！"}}
	}

	var events []event.Event
	if def.Effect.Cost > 0 {
		events = append(events, event.EventMP{Entity: actor, Amount: -def.Effect.Cost})
	}

	events = append(events, event.EventLog{Text: def.Effect.Message})
	pos := w.GetPosition(actor)

	w.EachUnit(func(id entity.ID) {
		if id == actor {
			return
		}
		uPos := w.GetPosition(id)
		if uPos == nil {
			return
		}
		dist := math.Sqrt(float64((pos.X-uPos.X)*(pos.X-uPos.X) + (pos.Y-uPos.Y)*(pos.Y-uPos.Y)))
		if dist <= float64(def.Effect.Range) {
			events = append(events, event.EventProjectile{
				StartX: float64(uPos.X), StartY: float64(uPos.Y),
				EndX: float64(uPos.X), EndY: float64(uPos.Y),
				TotalFrames: 10,
				ColorHex:    def.Effect.Color,
				IsFlash:     true,
			})

			atk := BuildCombatant(actor, w)
			defCombatant := BuildCombatant(id, w)
			res := combat.ResolveCombat(atk, defCombatant, combat.CombatTypeMagical, def.Effect.Damage, component.ElementFromString(def.Effect.Element), w.PlayerID(), w.RNG())

			for _, log := range res.Log {
				events = append(events, event.EventLog{Text: log})
			}
			events = append(events, event.EventAttack{
				Attacker: actor, Defender: id,
				Damage: res.Damage, Missed: res.Missed, Dir: pos.Dir,
			})
			if res.Killed {
				events = append(events, event.EventDeath{Entity: id, Name: defCombatant.Name})
				if res.XP > 0 {
					events = append(events, event.EventXP{Entity: actor, Amount: res.XP})
				}
			}
		}
	})
	return events
}

func (a *UseItemAction) handleHealFaith(actor entity.ID, def *content.ItemDef, w WorldAccess) []event.Event {
	stats := w.GetStats(actor)
	if stats == nil {
		return nil
	}
	amount := 5 + stats.Fai*2
	return []event.Event{
		event.EventHeal{Entity: actor, Amount: amount},
		event.EventLog{Text: def.Effect.Message},
	}
}

type DropItemAction struct {
	ItemIdx int
}

func (a *DropItemAction) Execute(actor entity.ID, w WorldAccess) []event.Event {
	inv := w.GetInventory(actor)
	pos := w.GetPosition(actor)
	reg := w.Registry()
	if inv == nil || pos == nil || a.ItemIdx < 0 || a.ItemIdx >= len(inv.Items) {
		return nil
	}
	entry := inv.Items[a.ItemIdx]
	def, _ := reg.GetItemDef(entry.DefID)

	var events []event.Event
	if entry.Equipped {
		events = append(events, event.EventEquip{Entity: actor, ItemIdx: a.ItemIdx})
	}
	events = append(events, event.EventDrop{
		Entity: actor,
		X:      pos.X,
		Y:      pos.Y,
		Items:  []component.ItemEntry{entry},
	})
	events = append(events, event.EventItemConsume{Entity: actor, ItemIdx: a.ItemIdx, Count: entry.Count})
	events = append(events, event.EventLog{Text: def.Name + "を捨てた。"})

	return events
}
