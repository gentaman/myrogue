package rules

import (
	"math"
	"math/rand"

	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
)

type EntityAccess interface {
	GetStats(id entity.ID) *component.Stats
	GetName(id entity.ID) string
	GetRewardXP(id entity.ID) int
	IsAlive(id entity.ID) bool
	Destroy(id entity.ID)
	GetInventory(id entity.ID) *component.Inventory
	GetPosition(id entity.ID) *component.Position
}

type Resolver struct {
	Access   EntityAccess
	PlayerID entity.ID
}

func (r *Resolver) ProcessDeath(ev action.EventDeath) []action.Event {
	var events []action.Event
	pos := r.Access.GetPosition(ev.Entity)
	inv := r.Access.GetInventory(ev.Entity)
	if pos != nil && inv != nil && len(inv.Items) > 0 {
		events = append(events, action.EventDrop{X: pos.X, Y: pos.Y, Items: inv.Items})
	}
	r.Access.Destroy(ev.Entity)
	return events
}

func (r *Resolver) ProcessXP(ev action.EventXP) []action.Event {
	stats := r.Access.GetStats(ev.Entity)
	if stats == nil || stats.Level >= 100 {
		return nil
	}
	stats.XP += ev.Amount
	var events []action.Event
	for stats.XP >= stats.XPToNext && stats.Level < 100 {
		stats.XP -= stats.XPToNext
		stats.Level++
		stats.XPToNext = int(10 * math.Pow(float64(stats.Level), 1.5))
		events = append(events, action.EventLevelUp{Entity: ev.Entity, NewLevel: stats.Level})
		rollStatsUp(stats, rand.New(rand.NewSource(rand.Int63())))
	}
	return events
}

func rollStatsUp(s *component.Stats, rng *rand.Rand) {
	count := rng.Intn(3) + 1
	for i := 0; i < count; i++ {
		stat := rng.Intn(6)
		switch stat {
		case 0:
			s.Str++
			if s.Str > component.MaxStr {
				s.Str = component.MaxStr
			}
		case 1:
			s.Wis++
			if s.Wis > component.MaxWis {
				s.Wis = component.MaxWis
			}
			s.MP += 5
			s.MaxMP += 5
		case 2:
			s.Fai++
			if s.Fai > component.MaxFai {
				s.Fai = component.MaxFai
			}
		case 3:
			s.Vit++
			if s.Vit > component.MaxVit {
				s.Vit = component.MaxVit
			}
			s.HP += 2
			s.MaxHP += 2
		case 4:
			s.Agi++
			if s.Agi > component.MaxAgi {
				s.Agi = component.MaxAgi
			}
		case 5:
			s.Luk++
			if s.Luk > component.MaxLuk {
				s.Luk = component.MaxLuk
			}
		}
	}
}
