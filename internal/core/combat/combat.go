package combat

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/rules"
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
	XP      int // Awarded when this combatant is killed
}

type CombatResult struct {
	AttackerID entity.ID
	DefenderID entity.ID
	Damage     int
	Missed     bool
	Killed     bool
	Log        []string
	XP         int
}

func TryHit(attacker, defender *Combatant, rng rules.RNG) bool {
	if defender.Stats.Luk > attacker.Stats.Luk {
		lukRate := float64(defender.Stats.Luk) / float64(component.MaxLuk)
		dodgeChance := int(lukRate * 90)
		if rng.Intn(100) < dodgeChance {
			return false
		}
	}
	chance := (attacker.Stats.Agi - defender.Stats.Agi) * 5
	attackChance := 100 + chance
	return rng.Intn(100) < attackChance
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

func ResolveCombat(attacker, defender *Combatant, cType CombatType, bonusAtk int, element component.Element, playerID entity.ID, rng rules.RNG) CombatResult {
	res := CombatResult{
		AttackerID: attacker.ID,
		DefenderID: defender.ID,
	}

	if !TryHit(attacker, defender, rng) {
		res.Missed = true
		res.Log = append(res.Log, fmt.Sprintf("%sは%sの攻撃をかわした！", defender.Name, attacker.Name))
		return res
	}

	damage := CalcDamage(attacker, defender, cType, bonusAtk, element)
	multiplier := component.GetEffectiveness(element, defender.Element)

	if multiplier > 1.1 {
		res.Log = append(res.Log, "効果は抜群だ！")
	} else if multiplier < 0.9 {
		res.Log = append(res.Log, "効果はいまひとつのようだ...")
	}

	if attacker.ID == playerID {
		res.Log = append(res.Log, fmt.Sprintf("%sに %d のダメージを与えた！", defender.Name, damage))
	} else if defender.ID == playerID {
		res.Log = append(res.Log, fmt.Sprintf("%sから %d のダメージを受けた！", attacker.Name, damage))
	} else {
		res.Log = append(res.Log, fmt.Sprintf("%sが%sに %d のダメージを与えた！", attacker.Name, defender.Name, damage))
	}

	defender.Stats.HP -= damage
	res.Damage = damage

	if defender.Stats.HP <= 0 {
		res.Killed = true
		if attacker.ID != defender.ID && defender.XP > 0 {
			res.XP = defender.XP
		}
	}

	return res
}
