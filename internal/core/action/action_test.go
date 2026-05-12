package action

import (
	"testing"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
	"github.com/gentaman/myrogue/internal/core/rules"
	"github.com/gentaman/myrogue/internal/core/world"
)

type mockWorld struct {
	pos     map[entity.ID]*component.Position
	stats   map[entity.ID]*component.Stats
	inv     map[entity.ID]*component.Inventory
	names   map[entity.ID]string
	gamemap *world.GameMap
	player  entity.ID
	rng     rules.RNG
	reg     *content.Registry
}

func (m *mockWorld) GetPosition(id entity.ID) *component.Position   { return m.pos[id] }
func (m *mockWorld) GetStats(id entity.ID) *component.Stats         { return m.stats[id] }
func (m *mockWorld) GetFaction(id entity.ID) *component.FactionComp { return nil }
func (m *mockWorld) GetInventory(id entity.ID) *component.Inventory { return m.inv[id] }
func (m *mockWorld) GetAI(id entity.ID) *component.AI               { return nil }
func (m *mockWorld) GetAnim(id entity.ID) *component.AnimState      { return nil }
func (m *mockWorld) GetName(id entity.ID) string                    { return m.names[id] }
func (m *mockWorld) GetElement(id entity.ID) component.Element      { return component.ElementNone }
func (m *mockWorld) GetRace(id entity.ID) component.Race            { return component.RaceHuman }
func (m *mockWorld) GetRewardXP(id entity.ID) int                   { return 10 }
func (m *mockWorld) IsAlive(id entity.ID) bool                      { return true }
func (m *mockWorld) UnitAt(x, y int) entity.ID {
	for id, p := range m.pos {
		if p.X == x && p.Y == y {
			return id
		}
	}
	return entity.InvalidID
}
func (m *mockWorld) Map() *world.GameMap         { return m.gamemap }
func (m *mockWorld) PlayerID() entity.ID         { return m.player }
func (m *mockWorld) RNG() rules.RNG              { return m.rng }
func (m *mockWorld) Registry() *content.Registry { return m.reg }

type mockRNG struct{ val int }

func (m *mockRNG) Intn(n int) int   { return m.val }
func (m *mockRNG) Float64() float64 { return 0 }

func TestMoveAction(t *testing.T) {
	gm := &world.GameMap{}
	for i := 0; i < world.MapWidth; i++ {
		for j := 0; j < world.MapHeight; j++ {
			gm.Tiles[i][j] = world.Floor
		}
	}

	w := &mockWorld{
		pos: map[entity.ID]*component.Position{
			entity.ID(1): {X: 1, Y: 1},
		},
		gamemap: gm,
	}

	act := &MoveAction{DX: 1, DY: 0}
	events := act.Execute(entity.ID(1), w)

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ev, ok := events[0].(event.EventMoved)
	if !ok {
		t.Fatalf("expected EventMoved, got %T", events[0])
	}
	if ev.ToX != 2 || ev.ToY != 1 {
		t.Errorf("expected move to (2, 1), got (%d, %d)", ev.ToX, ev.ToY)
	}
}

func TestAttackAction(t *testing.T) {
	w := &mockWorld{
		pos: map[entity.ID]*component.Position{
			entity.ID(1): {X: 1, Y: 1},
			entity.ID(2): {X: 2, Y: 1},
		},
		stats: map[entity.ID]*component.Stats{
			entity.ID(1): {HP: 10, Str: 5, Agi: 10, Luk: 10},
			entity.ID(2): {HP: 10, Vit: 2, Agi: 10, Luk: 10},
		},
		names: map[entity.ID]string{
			entity.ID(1): "Atk",
			entity.ID(2): "Def",
		},
		player: entity.ID(1),
		rng:    &mockRNG{val: 0},
		reg:    content.NewRegistry(),
	}

	act := &AttackAction{TargetX: 2, TargetY: 1}
	events := act.Execute(entity.ID(1), w)

	// Expected events: EventLog (damage), EventAttack
	hasAttack := false
	for _, ev := range events {
		if _, ok := ev.(event.EventAttack); ok {
			hasAttack = true
		}
	}
	if !hasAttack {
		t.Errorf("expected EventAttack in events")
	}
}
