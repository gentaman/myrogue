package action

import (
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
	gm := w.Map()
	if !gm.InBounds(nx, ny) || !gm.IsPassable(nx, ny) {
		return nil
	}
	pos.Dir = component.DirFromDelta(a.DX, a.DY)

	occupant := w.UnitAt(nx, ny)
	if occupant != entity.InvalidID {
		return nil
	}
	oldX, oldY := pos.X, pos.Y
	pos.X, pos.Y = nx, ny
	return []event.Event{event.EventMoved{Entity: actor, FromX: oldX, FromY: oldY, ToX: nx, ToY: ny}}
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
	pos.Dir = component.DirFromDelta(dx, dy)

	anim := w.GetAnim(actor)
	if anim != nil {
		anim.AttackAnim = component.AttackAnimFrames
	}

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
	stats := w.GetStats(id)
	if stats == nil {
		return &combat.Combatant{ID: id, Stats: &component.Stats{}}
	}
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
		Stats:   stats,
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
	return []event.Event{event.EventFloorChange{
		CurrentFloor: w.Map().Floor,
		Direction:    a.Direction,
	}}
}
