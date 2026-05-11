package combat

import (
	"fmt"
	"math/rand"

	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
)

type CombatType int

const (
	CombatTypePhysical CombatType = iota
	CombatTypeMagical
)

type Combatant struct {
	ID      entity.ID
	Name    string
	Stats   *component.Stats
	Element component.Element
	Race    component.Race
	PhyAtk  int
	PhyDef  int
	MagAtk  int
	MagDef  int
}

func TryHit(attacker, defender *Combatant) bool {
	if defender.Stats.Luk > attacker.Stats.Luk {
		lukRate := float64(defender.Stats.Luk) / float64(component.MaxLuk)
		dodgeChance := int(lukRate * 90)
		if rand.Intn(100) < dodgeChance {
			return false
		}
	}
	chance := (attacker.Stats.Agi - defender.Stats.Agi) * 5
	attackChance := 100 + chance
	return rand.Intn(100) < attackChance
}

func CalcDamage(attacker, defender *Combatant, cType CombatType, bonusAtk int, element component.Element) int {
	multiplier := component.GetEffectiveness(element, defender.Element)
	var attack, defence int
	switch cType {
	case CombatTypePhysical:
		attack = attacker.Stats.Str + attacker.PhyAtk + bonusAtk
		defence = defender.Stats.Vit + defender.PhyDef
	case CombatTypeMagical:
		attack = attacker.Stats.Wis + attacker.MagAtk + bonusAtk
		defence = defender.Stats.Fai + defender.MagDef
	}
	damage := int(float64(attack-defence) * multiplier)
	if damage < 0 {
		damage = 0
	}
	return damage
}

func ResolveCombat(attacker, defender *Combatant, cType CombatType, bonusAtk int, element component.Element, playerID entity.ID) []action.Event {
	var events []action.Event
	if !TryHit(attacker, defender) {
		events = append(events, action.EventLog{
			Text: fmt.Sprintf("%sは%sの攻撃をかわした！", defender.Name, attacker.Name),
		})
		return events
	}

	damage := CalcDamage(attacker, defender, cType, bonusAtk, element)
	multiplier := component.GetEffectiveness(element, defender.Element)

	if multiplier > 1.1 {
		events = append(events, action.EventLog{Text: "効果は抜群だ！"})
	} else if multiplier < 0.9 {
		events = append(events, action.EventLog{Text: "効果はいまひとつのようだ..."})
	}

	if attacker.ID == playerID {
		events = append(events, action.EventLog{Text: fmt.Sprintf("%sに %d のダメージを与えた！", defender.Name, damage)})
	} else if defender.ID == playerID {
		events = append(events, action.EventLog{Text: fmt.Sprintf("%sから %d のダメージを受けた！", attacker.Name, damage)})
	} else {
		events = append(events, action.EventLog{Text: fmt.Sprintf("%sが%sに %d のダメージを与えた！", attacker.Name, defender.Name, damage)})
	}

	defender.Stats.HP -= damage
	events = append(events, action.EventAttack{Attacker: attacker.ID, Defender: defender.ID, Damage: damage})

	if defender.Stats.HP <= 0 {
		events = append(events, action.EventDeath{Entity: defender.ID, Name: defender.Name})
	}
	if attacker.Stats.HP <= 0 {
		events = append(events, action.EventDeath{Entity: attacker.ID, Name: attacker.Name})
	}

	return events
}
