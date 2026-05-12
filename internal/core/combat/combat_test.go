package combat

import (
	"testing"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
)

type mockRNG struct {
	val int
}

func (m *mockRNG) Intn(n int) int   { return m.val }
func (m *mockRNG) Float64() float64 { return 0 }

func TestResolveCombat_Hit(t *testing.T) {
	atk := &Combatant{
		ID:    entity.ID(1),
		Name:  "Attacker",
		Stats: &component.Stats{HP: 10, MaxHP: 10, Str: 5, Agi: 10, Luk: 10},
	}
	def := &Combatant{
		ID:    entity.ID(2),
		Name:  "Defender",
		Stats: &component.Stats{HP: 10, MaxHP: 10, Vit: 2, Agi: 10, Luk: 10},
	}
	rng := &mockRNG{val: 50} // 100 + (10-10)*5 = 100. 50 < 100 is hit.

	res := ResolveCombat(atk, def, CombatTypePhysical, 0, component.ElementNone, entity.ID(1), rng)

	if res.Missed {
		t.Errorf("expected hit, got miss")
	}
	if res.Damage != 3 { // 5 - 2 = 3
		t.Errorf("expected 3 damage, got %d", res.Damage)
	}
	if def.Stats.HP != 7 {
		t.Errorf("expected defender HP 7, got %d", def.Stats.HP)
	}
}

func TestResolveCombat_Miss(t *testing.T) {
	atk := &Combatant{
		ID:    entity.ID(1),
		Name:  "Attacker",
		Stats: &component.Stats{HP: 10, MaxHP: 10, Str: 5, Agi: 10, Luk: 10},
	}
	def := &Combatant{
		ID:    entity.ID(2),
		Name:  "Defender",
		Stats: &component.Stats{HP: 10, MaxHP: 10, Vit: 2, Agi: 10, Luk: 10},
	}
	rng := &mockRNG{val: 150} // 100 + (10-10)*5 = 100. 150 < 100 is false.

	res := ResolveCombat(atk, def, CombatTypePhysical, 0, component.ElementNone, entity.ID(1), rng)

	if !res.Missed {
		t.Errorf("expected miss, got hit")
	}
}

func TestResolveCombat_Kill(t *testing.T) {
	atk := &Combatant{
		ID:    entity.ID(1),
		Name:  "Attacker",
		Stats: &component.Stats{HP: 10, MaxHP: 10, Str: 20, Agi: 10, Luk: 10},
	}
	def := &Combatant{
		ID:    entity.ID(2),
		Name:  "Defender",
		Stats: &component.Stats{HP: 5, MaxHP: 5, Vit: 2, Agi: 10, Luk: 10},
		XP:    100,
	}
	rng := &mockRNG{val: 0}

	res := ResolveCombat(atk, def, CombatTypePhysical, 0, component.ElementNone, entity.ID(1), rng)

	if !res.Killed {
		t.Errorf("expected kill, got no kill")
	}
	if res.XP != 100 {
		t.Errorf("expected 100 XP, got %d", res.XP)
	}
}
