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
	Dir          component.Dir
}

func (EventMoved) isEvent() {}

type EventAttack struct {
	Attacker entity.ID
	Defender entity.ID
	Damage   int
	Missed   bool
	Dir      component.Dir
}

func (EventAttack) isEvent() {}

type EventDeath struct {
	Entity entity.ID
	Name   string
}

func (EventDeath) isEvent() {}

type EventDirChanged struct {
	Entity entity.ID
	Dir    component.Dir
}

func (EventDirChanged) isEvent() {}

type EventXP struct {
	Entity entity.ID
	Amount int
}

func (EventXP) isEvent() {}

type EventMP struct {
	Entity entity.ID
	Amount int // Positive for gain, negative for cost
}

func (EventMP) isEvent() {}

type EventHeal struct {
	Entity entity.ID
	Amount int
}

func (EventHeal) isEvent() {}

type EventStatusEffect struct {
	Entity   entity.ID
	Effect   component.StatusType
	Duration int
}

func (EventStatusEffect) isEvent() {}

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
	Entity entity.ID
	X, Y   int
	Items  []component.ItemEntry
}

func (EventDrop) isEvent() {}

type EventEquip struct {
	Entity  entity.ID
	ItemIdx int
}

func (EventEquip) isEvent() {}

type EventItemConsume struct {
	Entity     entity.ID
	ItemIdx    int
	Count      int
	Durability int
}

func (EventItemConsume) isEvent() {}

type EventProjectile struct {
	StartX, StartY float64
	EndX, EndY     float64
	TotalFrames    int
	ColorHex       string
	IsFlash        bool
}

func (EventProjectile) isEvent() {}

type EventVisual struct {
	ID               string
	SourceX, SourceY float64
	TargetX, TargetY float64
	ColorHex         string // Optional override
}

func (EventVisual) isEvent() {}

type EventWin struct {
	Score int
}

func (EventWin) isEvent() {}

type EventRevealMap struct{}

func (EventRevealMap) isEvent() {}

type EventSwap struct {
	Entity1, Entity2 entity.ID
}

func (EventSwap) isEvent() {}

type EventWait struct {
	Entity entity.ID
}

func (EventWait) isEvent() {}
