package action

import (
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
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
}

type Action interface {
	Execute(actor entity.ID, w WorldAccess) []Event
}

type MoveAction struct {
	DX, DY int
}

func (a *MoveAction) Execute(actor entity.ID, w WorldAccess) []Event {
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
	return []Event{EventMoved{Entity: actor, FromX: oldX, FromY: oldY, ToX: nx, ToY: ny}}
}

type AttackAction struct {
	TargetX, TargetY int
}

func (a *AttackAction) Execute(actor entity.ID, w WorldAccess) []Event {
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
		return []Event{EventLog{Text: "何もない方向に攻撃した。"}}
	}
	return []Event{EventAttack{Attacker: actor, Defender: target}}
}

type WaitAction struct{}

func (a *WaitAction) Execute(actor entity.ID, _ WorldAccess) []Event {
	return []Event{EventWait{Entity: actor}}
}
