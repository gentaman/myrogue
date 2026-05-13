package rules

import (
	"math"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
)

type EntityAccess interface {
	GetStats(id entity.ID) *component.Stats
	GetName(id entity.ID) string
	GetRewardXP(id entity.ID) int
	IsAlive(id entity.ID) bool
	Destroy(id entity.ID)
	GetInventory(id entity.ID) *component.Inventory
	GetPosition(id entity.ID) *component.Position
	ApplyStatus(id entity.ID, effect component.StatusType, duration int)
	Registry() *content.Registry
}

type Resolver struct {
	Access   EntityAccess
	PlayerID entity.ID
	RNG      RNG
}

func (r *Resolver) ProcessDeath(ev event.EventDeath) []event.Event {
	var events []event.Event
	pos := r.Access.GetPosition(ev.Entity)
	inv := r.Access.GetInventory(ev.Entity)
	if pos != nil && inv != nil && len(inv.Items) > 0 {
		events = append(events, event.EventDrop{Entity: ev.Entity, X: pos.X, Y: pos.Y, Items: inv.Items})
	}
	r.Access.Destroy(ev.Entity)
	return events
}

func (r *Resolver) ProcessEquip(ev event.EventEquip) []event.Event {
	inv := r.Access.GetInventory(ev.Entity)
	reg := r.Access.Registry()
	if inv == nil || reg == nil || ev.ItemIdx < 0 || ev.ItemIdx >= len(inv.Items) {
		return nil
	}

	entry := &inv.Items[ev.ItemIdx]
	def, ok := reg.GetItemDef(entry.DefID)
	if !ok {
		return nil
	}

	entry.Equipped = !entry.Equipped
	if entry.Equipped {
		// unequip others in same slot
		for i := range inv.Items {
			if i != ev.ItemIdx && inv.Items[i].Equipped {
				otherDef, ok2 := reg.GetItemDef(inv.Items[i].DefID)
				if ok2 && otherDef.EquipSlot == def.EquipSlot {
					inv.Items[i].Equipped = false
				}
			}
		}
	}
	return nil
}

func (r *Resolver) ProcessItemConsume(ev event.EventItemConsume) []event.Event {
	inv := r.Access.GetInventory(ev.Entity)
	if inv == nil || ev.ItemIdx < 0 || ev.ItemIdx >= len(inv.Items) {
		return nil
	}
	entry := &inv.Items[ev.ItemIdx]
	if ev.Durability > 0 && entry.Durability > 0 {
		entry.Durability -= ev.Durability
		if entry.Durability <= 0 {
			entry.Count--
			// リセット耐久度は定義から取るべきだが、一旦簡易的に
			// 実際には壊れるパターンが多い
			if entry.Count > 0 {
				reg := r.Access.Registry()
				if reg != nil {
					if def, ok := reg.GetItemDef(entry.DefID); ok {
						entry.Durability = def.Durability
					}
				}
			}
		}
	} else if ev.Count > 0 {
		entry.Count -= ev.Count
	}

	if entry.Count <= 0 {
		inv.Items = append(inv.Items[:ev.ItemIdx], inv.Items[ev.ItemIdx+1:]...)
	}
	return nil
}

func (r *Resolver) ProcessAttack(ev event.EventAttack) []event.Event {
	if ev.Missed {
		return nil
	}
	stats := r.Access.GetStats(ev.Defender)
	if stats != nil {
		stats.HP -= ev.Damage
	}
	return nil
}

func (r *Resolver) ProcessMP(ev event.EventMP) []event.Event {
	stats := r.Access.GetStats(ev.Entity)
	if stats != nil {
		stats.MP += ev.Amount
		if stats.MP < 0 {
			stats.MP = 0
		}
		if stats.MP > stats.MaxMP {
			stats.MP = stats.MaxMP
		}
	}
	return nil
}

func (r *Resolver) ProcessHeal(ev event.EventHeal) []event.Event {
	stats := r.Access.GetStats(ev.Entity)
	if stats != nil {
		stats.HP += ev.Amount
		if stats.HP > stats.MaxHP {
			stats.HP = stats.MaxHP
		}
	}
	return nil
}

func (r *Resolver) ProcessStatusEffect(ev event.EventStatusEffect) []event.Event {
	r.Access.ApplyStatus(ev.Entity, ev.Effect, ev.Duration)
	return nil
}

func (r *Resolver) ProcessXP(ev event.EventXP) []event.Event {
	stats := r.Access.GetStats(ev.Entity)
	if stats == nil || stats.Level >= 100 {
		return nil
	}
	stats.XP += ev.Amount
	var events []event.Event
	for stats.XP >= stats.XPToNext && stats.Level < 100 {
		stats.XP -= stats.XPToNext
		stats.Level++
		stats.XPToNext = int(10 * math.Pow(float64(stats.Level), 1.5))
		events = append(events, event.EventLevelUp{Entity: ev.Entity, NewLevel: stats.Level})
		rollStatsUp(stats, r.RNG)
	}
	return events
}

func rollStatsUp(s *component.Stats, rng RNG) {
	if rng == nil {
		return
	}
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
