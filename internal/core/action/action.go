package action

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/combat"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
	"github.com/gentaman/myrogue/internal/core/rules"
	"github.com/gentaman/myrogue/internal/core/world"
)

type WorldAccess interface {
	GetPosition(id entity.ID) *component.Position
	GetStats(id entity.ID) *component.Stats
	GetFaction(id entity.ID) *component.FactionComp
	GetInventory(id entity.ID) *component.Inventory
	GetAI(id entity.ID) *component.AI
	GetAnim(id entity.ID) *component.AnimState
	GetName(id entity.ID) string
	GetElement(id entity.ID) component.Element
	GetRace(id entity.ID) component.Race
	GetRewardXP(id entity.ID) int
	IsAlive(id entity.ID) bool
	UnitAt(x, y int) entity.ID
	Map() *world.GameMap
	PlayerID() entity.ID
	RNG() rules.RNG
	Registry() *content.Registry
}

type Action interface {
	Execute(actor entity.ID, w WorldAccess) []event.Event
}

type MoveAction struct {
	DX, DY int
}

func (a *MoveAction) Execute(actor entity.ID, w WorldAccess) []event.Event {
	pos := w.GetPosition(actor)
	if pos == nil {
		return nil
	}
	nx, ny := pos.X+a.DX, pos.Y+a.DY
	dir := component.DirFromDelta(a.DX, a.DY)

	gm := w.Map()
	if !gm.InBounds(nx, ny) || !gm.IsPassable(nx, ny) {
		return []event.Event{event.EventDirChanged{Entity: actor, Dir: dir}}
	}

	occupant := w.UnitAt(nx, ny)
	if occupant != entity.InvalidID {
		return []event.Event{event.EventDirChanged{Entity: actor, Dir: dir}}
	}
	return []event.Event{event.EventMoved{
		Entity: actor,
		FromX:  pos.X,
		FromY:  pos.Y,
		ToX:    nx,
		ToY:    ny,
		Dir:    dir,
	}}
}

type AttackAction struct {
	TargetX, TargetY int
}

func (a *AttackAction) Execute(actor entity.ID, w WorldAccess) []event.Event {
	pos := w.GetPosition(actor)
	if pos == nil {
		return nil
	}
	dx := a.TargetX - pos.X
	dy := a.TargetY - pos.Y
	dir := component.DirFromDelta(dx, dy)

	target := w.UnitAt(a.TargetX, a.TargetY)
	if target == entity.InvalidID {
		return []event.Event{event.EventLog{Text: "何もない方向に攻撃した。"}}
	}

	atk := BuildCombatant(actor, w)
	def := BuildCombatant(target, w)

	res := combat.ResolveCombat(atk, def, combat.CombatTypePhysical, 0, atk.Element, w.PlayerID(), w.RNG())

	var events []event.Event
	for _, log := range res.Log {
		events = append(events, event.EventLog{Text: log})
	}

	events = append(events, event.EventAttack{
		Attacker: res.AttackerID,
		Defender: res.DefenderID,
		Damage:   res.Damage,
		Missed:   res.Missed,
		Dir:      dir,
	})

	if res.Killed {
		events = append(events, event.EventDeath{
			Entity: res.DefenderID,
			Name:   def.Name,
		})
		if res.XP > 0 {
			events = append(events, event.EventXP{
				Entity: res.AttackerID,
				Amount: res.XP,
			})
		}
	}

	return events
}

func BuildCombatant(id entity.ID, w WorldAccess) *combat.Combatant {
	statsPtr := w.GetStats(id)
	if statsPtr == nil {
		return &combat.Combatant{ID: id, Stats: &component.Stats{}}
	}
	// コピーを作成して副作用を防ぐ
	stats := *statsPtr

	inv := w.GetInventory(id)
	reg := w.Registry()

	phyAtk, phyDef, magAtk, magDef := 0, 0, 0, 0
	if inv != nil && reg != nil {
		for _, e := range inv.Items {
			if e.Equipped {
				if def, ok := reg.GetItemDef(e.DefID); ok {
					phyAtk += def.PhyAtk
					phyDef += def.PhyDef
					magAtk += def.MagAtk
					magDef += def.MagDef
				}
			}
		}
	}

	return &combat.Combatant{
		ID:      id,
		Name:    w.GetName(id),
		Stats:   &stats,
		Element: w.GetElement(id),
		Race:    w.GetRace(id),
		PhyAtk:  phyAtk,
		PhyDef:  phyDef,
		MagAtk:  magAtk,
		MagDef:  magDef,
		XP:      w.GetRewardXP(id),
	}
}

type WaitAction struct{}

func (a *WaitAction) Execute(actor entity.ID, _ WorldAccess) []event.Event {
	return []event.Event{event.EventWait{Entity: actor}}
}

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
		default:
			events = append(events, event.EventLog{Text: def.Name + "を使った。"})
		}
	} else {
		events = append(events, event.EventLog{Text: def.Name + "を使った。"})
	}

	events = append(events, event.EventItemConsume{Entity: actor, ItemIdx: a.ItemIdx, Count: 1})
	return events
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

type SkillAction struct {
	SkillID          string
	TargetX, TargetY int
}

func (a *SkillAction) Execute(actor entity.ID, w WorldAccess) []event.Event {
	reg := w.Registry()
	if reg == nil {
		return nil
	}
	sk, ok := reg.GetSkillDef(a.SkillID)
	if !ok {
		return nil
	}

	stats := w.GetStats(actor)
	if stats == nil || stats.MP < sk.MPCost {
		return []event.Event{event.EventLog{Text: "MPが足りない！"}}
	}

	var events []event.Event
	events = append(events, event.EventMP{Entity: actor, Amount: -sk.MPCost})

	switch sk.Type {
	case content.SkillTypeAttack:
		target := w.UnitAt(a.TargetX, a.TargetY)
		if target == entity.InvalidID {
			events = append(events, event.EventLog{Text: sk.Name + "を放った！ しかし何もいなかった。"})
			return events
		}

		pos := w.GetPosition(actor)
		if pos != nil && sk.Range > 1 {
			events = append(events, event.EventProjectile{
				StartX:      float64(pos.X),
				StartY:      float64(pos.Y),
				EndX:        float64(a.TargetX),
				EndY:        float64(a.TargetY),
				TotalFrames: 12,
				ColorHex:    sk.ColorHex,
			})
		}

		atk := BuildCombatant(actor, w)
		def := BuildCombatant(target, w)
		damage := combat.CalcDamage(atk, def, combat.CombatTypeMagical, sk.Power, sk.Element)
		if damage < 1 {
			damage = 1
		}

		events = append(events, event.EventLog{Text: fmt.Sprintf("%sで%sに %d ダメージ！", sk.Name, w.GetName(target), damage)})
		events = append(events, event.EventAttack{
			Attacker: actor,
			Defender: target,
			Damage:   damage,
			Dir:      component.DirFromDelta(a.TargetX-pos.X, a.TargetY-pos.Y),
		})

		if def.Stats.HP-damage <= 0 {
			events = append(events, event.EventDeath{Entity: target, Name: def.Name})
			if def.XP > 0 {
				events = append(events, event.EventXP{Entity: actor, Amount: def.XP})
			}
		}

	case content.SkillTypeHeal:
		heal := sk.Power + stats.Wis/2
		events = append(events, event.EventHeal{Entity: actor, Amount: heal})
		events = append(events, event.EventLog{Text: fmt.Sprintf("%sを使った！ HPが%d回復した。", sk.Name, heal)})

	case content.SkillTypeBuff:
		if sk.Status != "" {
			events = append(events, event.EventStatusEffect{
				Entity:   actor,
				Effect:   component.StatusFromString(sk.Status),
				Duration: sk.Duration,
			})
		}
		events = append(events, event.EventLog{Text: sk.Name + "を使った！"})

	case content.SkillTypeDebuff:
		target := w.UnitAt(a.TargetX, a.TargetY)
		if target != entity.InvalidID && sk.Status != "" {
			events = append(events, event.EventStatusEffect{
				Entity:   target,
				Effect:   component.StatusFromString(sk.Status),
				Duration: sk.Duration,
			})
			events = append(events, event.EventLog{Text: fmt.Sprintf("%sに%sをかけた！", w.GetName(target), sk.Name)})
		} else {
			events = append(events, event.EventLog{Text: sk.Name + "を放った！ しかし何もいなかった。"})
		}
	}

	return events
}
