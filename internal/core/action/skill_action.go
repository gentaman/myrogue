package action

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/combat"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/event"
)

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
