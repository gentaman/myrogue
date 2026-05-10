package game

import (
	"fmt"
	"math/rand"
)

type Element int

const (
	ElementNone Element = iota
	ElementFire
	ElementWater
	ElementAir
	ElementEarth
	ElementLight
	ElementDark
)

type Race int

const (
	RaceHuman Race = iota
	RaceElf
	RaceDwarf
	RaceGnome
	RaceHalfling
	RaceElement
	RaceBeast
	RaceDragon
	RacePlant
	RaceUndead
	RaceInsect
	RaceBird
	RaceDemon
	RaceMachine
	RaceHolyBeast
)

var RaceNames = map[Race]string{
	RaceHuman:     "ヒューマン",
	RaceElf:       "エルフ",
	RaceDwarf:     "ドワーフ",
	RaceGnome:     "ノーム",
	RaceHalfling:  "ハーフリング",
	RaceElement:   "エレメント",
	RaceBeast:     "野獣",
	RaceDragon:    "ドラゴン",
	RacePlant:     "植物",
	RaceUndead:    "アンデッド",
	RaceInsect:    "虫",
	RaceBird:      "鳥",
	RaceDemon:     "悪魔",
	RaceMachine:   "機械",
	RaceHolyBeast: "聖獣",
}

var OrderedRaces = []Race{
	RaceHuman, RaceElf, RaceDwarf, RaceGnome, RaceHalfling,
	RaceElement, RaceBeast, RaceDragon, RacePlant, RaceUndead,
	RaceInsect, RaceBird, RaceDemon, RaceMachine, RaceHolyBeast,
}

type RelationState int

type CombatType int

const (
	CombatTypePhysical CombatType = iota
	CombatTypeMagical
	CombatTypeKarma
)

const (
	RelationHostile RelationState = iota
	RelationNeutral
	RelationFriendly
)

type CombatManager struct {
}

func tryAttack(attacker, defender Battler) bool {
	atkStats := attacker.GetStats()
	defStats := defender.GetStats()

	// 防御側の運が高ければ、運で避けられる
	// 0 - 90%の確率で避けられる
	if defStats.Luk > atkStats.Luk {
		lukRate := float64(defStats.Luk) / float64(MaxLuk)
		dodgeChance := int(lukRate * 90)
		if rand.Intn(100) < dodgeChance {
			return true
		}
	}

	// Agiの差が 5%ずつ影響する
	chance := (atkStats.Agi - defStats.Agi) * 5

	attackChance := atkStats.BaseAccuracy + chance

	return rand.Intn(100) < int(attackChance*100)
}

func (cm *CombatManager) ResolveCombat(bus *EventBus, attacker, defender Battler, cType CombatType, bonusAtk int, element Element) {
	if tryAttack(attacker, defender) {
		cm.ExecuteAttack(bus, cType, attacker, defender, bonusAtk, element)
	} else {
		bus.Publish(MsgLog{Text: fmt.Sprintf("%sは%sの攻撃をかわした！", defender.GetName(), attacker.GetName())})
	}

	cm.HandlePostCombat(bus, attacker, defender)
}

func (cm *CombatManager) HandlePostCombat(bus *EventBus, attacker, defender Battler) {
	toXPatk := -1
	toXPdef := -1
	if defender.GetCurrentHP() <= 0 {
		OnDeath(bus, defender)
		toXPatk = defender.GetRewardXP()
	}
	if attacker.GetCurrentHP() <= 0 {
		OnDeath(bus, attacker)
		toXPdef = attacker.GetRewardXP()
	}

	if toXPatk >= 0 {
		attacker.GainXP(bus, toXPatk)
	}
	if toXPdef >= 0 {
		defender.GainXP(bus, toXPdef)
	}
}

func OnDeath(bus *EventBus, battler Battler) {
	battler.OnDeath(bus)
	bus.Publish(MsgDeath{Battler: battler})
}

// 戦闘に参加する存在のインターフェース
type Battler interface {
	GetID() int64                                  // ユニークID
	GetStats() Stats                               // 攻撃力、属性などを返す
	ApplyDamage(dmg int)                           // ダメージを受ける処理
	GetName() string                               // 名前（ログ用）
	GetRace() Race                                 // 種族
	GetLevel() int                                 // レベル
	ConsumeDurability(bus *EventBus, et EquipType) // 装備の耐久度を減らす
	UpdateRelation(targetID int64, delta int)      // 関係性の更新
	GetRelation(target Battler) RelationState      // 対象への関係性
	GetCombatType() CombatType                     // 攻撃側の戦闘形式
	GetCurrentHP() int                             // 現在のHP
	GetCurrentMP() int                             // 現在のMP
	GetRewardXP() int                              // 報酬経験値
	GainXP(bus *EventBus, amount int)              // 経験値の取得処理
	OnDeath(bus *EventBus)                         // 死亡時の処理
}

type Stats struct {
	PhysicalAttack               int
	PhysicalDefense              int
	MagicalAttack                int
	MagicalDefense               int
	BaseAccuracy                 int // 0-100の間
	Element                      Element
	Str, Wis, Fai, Vit, Agi, Luk int
}

// 相性倍率の定義（例: 2.0倍、0.5倍など）
var elementEffectiveness = map[Element]map[Element]float64{
	ElementFire: {
		ElementAir: 2.0, // 火は風に強い
	},
	ElementWater: {
		ElementFire: 2.0, // 水は火に強い
	},
	ElementAir: {
		ElementEarth: 2.0, // 風は土に強い
	},
	ElementEarth: {
		ElementWater: 2.0, // 土は水に強い
	},
	ElementLight: {
		ElementDark: 2.0, // 光は闇に強い
	},
	ElementDark: {
		ElementLight: 2.0, // 闇は光に強い
	},
}

// 相性倍率を取得するヘルパー関数
func GetEffectiveness(attacker, defender Element) float64 {
	if defs, ok := elementEffectiveness[attacker]; ok {
		if rate, ok := defs[defender]; ok {
			return rate
		}
	}
	return 1.0 // 通常ダメージ
}

func (cm *CombatManager) ExecuteAttack(bus *EventBus, cType CombatType, attacker Battler, defender Battler, bonusAtk int, element Element) int {
	atkStats := attacker.GetStats()
	defStats := defender.GetStats()

	// 1. ダメージ計算 (属性相性も適用)
	multiplier := GetEffectiveness(element, defStats.Element)

	attack := 0
	defence := 0

	switch cType {
	case CombatTypePhysical:
		{
			attack += atkStats.PhysicalAttack + bonusAtk
			defence += defStats.PhysicalDefense
		}
	case CombatTypeMagical:
		{
			attack += atkStats.MagicalAttack + bonusAtk
			defence += defStats.MagicalDefense
		}
	}

	damage := int(float64(attack-defence) * multiplier)

	if damage < 0 {
		damage = 0 // 最低0ダメージ
	}

	// 2. 関係性の更新 (攻撃された側は攻撃した側を嫌う)
	// 0ダメージでも嫌う
	defender.UpdateRelation(attacker.GetID(), min(-damage, -1))

	// 3. 耐久度消費
	if cType == CombatTypePhysical {
		attacker.ConsumeDurability(bus, EquipWeapon)
	}
	defender.ConsumeDurability(bus, EquipShield)
	defender.ConsumeDurability(bus, EquipArmor)

	// 4. ログ出力
	if multiplier > 1.1 {
		bus.Publish(MsgLog{Text: "効果は抜群だ！"})
	} else if multiplier < 0.9 {
		bus.Publish(MsgLog{Text: "効果はいまひとつのようだ..."})
	}

	if attacker.GetID() == 1 { // TODO: プレイヤーIDの定数化
		bus.Publish(MsgLog{Text: fmt.Sprintf("%sに %d のダメージを与えた！", defender.GetName(), damage)})
	} else if defender.GetID() == 1 {
		bus.Publish(MsgLog{Text: fmt.Sprintf("%sから %d のダメージを受けた！", attacker.GetName(), damage)})
	} else {
		bus.Publish(MsgLog{Text: fmt.Sprintf("%sが%sに %d のダメージを与えた！", attacker.GetName(), defender.GetName(), damage)})
	}

	// 5. 実行
	defender.ApplyDamage(damage)
	bus.Publish(MsgDamage{AttackerID: attacker.GetID(), TargetID: defender.GetID(), Damage: damage})

	return damage
}
