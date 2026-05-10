package game

import (
	"fmt"
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

const (
	RelationHostile RelationState = iota
	RelationNeutral
	RelationFriendly
)

type CombatManager struct {
}

func (cm *CombatManager) AttackEnemy(g *GameScene, enemyIdx int) {
	if enemyIdx < 0 || enemyIdx >= len(g.enemies) {
		return
	}
	enemy := &g.enemies[enemyIdx]

	cm.ExecuteAttack(g, g, enemy)

	if enemy.HP <= 0 {
		g.actorGainXP(&g.Player, enemyDefs[enemy.kind].xp) // あなたが経験値を獲得
		g.dropChest(enemy.Actor.X, enemy.Actor.Y, enemy.Actor.Inventory)
		g.enemies = append(g.enemies[:enemyIdx], g.enemies[enemyIdx+1:]...)
	}
}

func (cm *CombatManager) AttackPlayer(g *GameScene, enemyIdx int) {
	if enemyIdx < 0 || enemyIdx >= len(g.enemies) {
		return
	}
	enemy := &g.enemies[enemyIdx]

	// 命中・回避判定
	if g.tryDodge(enemyDefs[enemy.kind].acc) {
		g.pushMessage(fmt.Sprintf("%sの攻撃をかわした！", enemy.GetName()))
		return
	}

	cm.ExecuteAttack(g, enemy, g)

	if g.Player.HP <= 0 {
		g.Player.HP = 0
		g.playState = StateDead
		g.pushMessage("あなたは力尽きた...")
	}
}

// 戦闘に参加する存在のインターフェース
type Battler interface {
	GetID() int64                                 // ユニークID
	GetStats() Stats                              // 攻撃力、属性などを返す
	ApplyDamage(dmg int)                          // ダメージを受ける処理
	GetName() string                              // 名前（ログ用）
	GetRace() Race                                // 種族
	GetLevel() int                                // レベル
	ConsumeDurability(g *GameScene, et EquipType) // 装備の耐久度を減らす
	UpdateRelation(targetID int64, delta int)     // 関係性の更新
	GetRelation(target Battler) RelationState     // 対象への関係性
}

type Stats struct {
	Attack  int
	Defense int
	Element Element
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

func (cm *CombatManager) ExecuteAttack(g *GameScene, attacker Battler, defender Battler) int {
	atkStats := attacker.GetStats()
	defStats := defender.GetStats()

	// 1. ダメージ計算 (属性相性も適用)
	multiplier := GetEffectiveness(atkStats.Element, defStats.Element)
	damage := int(float64(atkStats.Attack-defStats.Defense) * multiplier)
	if damage < 1 {
		damage = 1 // 最低1ダメージ
	}

	// 2. 実行
	defender.ApplyDamage(damage)

	// 3. 関係性の更新 (攻撃された側は攻撃した側を嫌う)
	// ダメージ 1 につき -5 程度の悪化（閾値にあわせる）
	defender.UpdateRelation(attacker.GetID(), -damage*5)

	// 4. 耐久度消費
	attacker.ConsumeDurability(g, EquipWeapon)
	defender.ConsumeDurability(g, EquipShield)
	defender.ConsumeDurability(g, EquipArmor)

	// 5. ログ出力
	if multiplier > 1.1 {
		g.pushMessage("効果は抜群だ！")
	} else if multiplier < 0.9 {
		g.pushMessage("効果はいまひとつのようだ...")
	}

	if attacker.GetID() == g.Player.ID {
		g.pushMessage(fmt.Sprintf("%sに %d のダメージを与えた！", defender.GetName(), damage))
	} else if defender.GetID() == g.Player.ID {
		g.pushMessage(fmt.Sprintf("%sから %d のダメージを受けた！", attacker.GetName(), damage))
	} else {
		g.pushMessage(fmt.Sprintf("%sが%sに %d のダメージを与えた！", attacker.GetName(), defender.GetName(), damage))
	}

	return damage
}
