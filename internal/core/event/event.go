package event

import (
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
)

type Event interface{ isEvent() }

type EventMoved struct {
	Entity       entity.ID
	FromX, FromY int
	ToX, ToY     int
}

func (EventMoved) isEvent() {}

type EventAttack struct {
	Attacker entity.ID
	Defender entity.ID
	Damage   int
	Missed   bool
}

func (EventAttack) isEvent() {}

type EventDeath struct {
	Entity entity.ID
	Name   string
}

func (EventDeath) isEvent() {}

type EventXP struct {
	Entity entity.ID
	Amount int
}

func (EventXP) isEvent() {}

type EventLevelUp struct {
	Entity   entity.ID
	NewLevel int
}

func (EventLevelUp) isEvent() {}

type EventLog struct {
	Text string
}

func (EventLog) isEvent() {}

type EventSFX struct {
	ID string
}

func (EventSFX) isEvent() {}

type EventFloorChange struct {
	CurrentFloor int
	Direction    int
}

func (EventFloorChange) isEvent() {}

type EventDrop struct {
	X, Y  int
	Items []component.ItemEntry
}

func (EventDrop) isEvent() {}

type EventProjectile struct {
	StartX, StartY float64
	EndX, EndY     float64
	TotalFrames    int
	ColorHex       string
	IsFlash        bool
}

func (EventProjectile) isEvent() {}

type EventSwap struct {
	Entity1, Entity2 entity.ID
}

func (EventSwap) isEvent() {}

type EventWait struct {
	Entity entity.ID
}

func (EventWait) isEvent() {}
