package game

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// ItemKind はアイテムの種別を表す
type ItemKind int

// アイテムの定数定義（互換性のために残すが、基本的には JSON の順序に依存）
const (
	ItemHealSmall ItemKind = iota
	ItemHealLarge
	ItemRevealMap
	ItemClairvoyant
)
// ItemEffect はアイテム使用時の効果コマンド
type ItemEffect func(g *GameScene, entryIdx int)

type EquipType int

const (
	EquipNone EquipType = iota
	EquipWeapon
	EquipShield
	EquipArmor
)

type itemDef struct {
	name       string
	desc       string
	effectDesc string
	weight     int
	rarity     int // 大きいほど希少である。出現率に影響する
	clr        color.RGBA
	// floorBound が true のとき、取得フロアをスタック単位として管理し使用制限をかける
	floorBound bool
	// canUse が nil でなければ使用前に呼び出し、false のとき使用不可メッセージを設定する
	canUse func(g *GameScene, entry InventoryEntry) (bool, string)
	effect ItemEffect

	equipType EquipType
	atk       int
	def       int
}

// entryName はインベントリ行に表示する名前を返す（floorBound アイテムはフロア番号を付加、装備中は [E] を付加）
func (d *itemDef) entryName(entry InventoryEntry) string {
	name := d.name
	if d.floorBound {
		name = fmt.Sprintf("%s[F%d]", d.name, entry.obtainedFloor+1)
	}
	if entry.Equipped {
		name = "[E]" + name
	}
	return name
}

var (
	//go:embed assets/items.json
	itemsJSON []byte
	//go:embed assets/player.json
	playerJSON []byte

	itemDefs []itemDef

	itemIDMap = map[string]ItemKind{}

	playerMaxHP    int
	maxCarryWeight int
)

// parameterizedEffectSpec は JSON 内の効果定義
type parameterizedEffectSpec struct {
	Type      string `json:"type"`
	Amount    int    `json:"amount"`
	Message   string `json:"message"`
	EquipType string `json:"equip_type"`
	Atk       int    `json:"atk"`
	Def       int    `json:"def"`
	Cost      int    `json:"cost"`
	Damage    int    `json:"damage"`
}

// parameterizedConditionSpec は JSON 内の条件定義
type parameterizedConditionSpec struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type rawItem struct {
	ID         string                     `json:"id"`
	Name       string                     `json:"name"`
	Desc       string                     `json:"desc"`
	EffectDesc string                     `json:"effect_desc"`
	Weight     int                        `json:"weight"`
	Rarity     int                        `json:"rarity"`
	Color      string                     `json:"color"`
	FloorBound bool                       `json:"floor_bound"`
	CanUse     *parameterizedConditionSpec `json:"can_use"`
	Effect     *parameterizedEffectSpec    `json:"effect"`
}

func init() {
	// プレイヤー設定ロード
	var playerCfg struct {
		MaxHP          int `json:"max_hp"`
		MaxCarryWeight int `json:"max_carry_weight"`
	}
	if err := json.Unmarshal(playerJSON, &playerCfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal player.json: %v", err))
	}
	playerMaxHP = playerCfg.MaxHP
	maxCarryWeight = playerCfg.MaxCarryWeight

	// アイテム設定ロード
	var rawItems []rawItem
	if err := json.Unmarshal(itemsJSON, &rawItems); err != nil {
		panic(fmt.Sprintf("failed to unmarshal items.json: %v", err))
	}

	itemDefs = make([]itemDef, len(rawItems))
	for i, raw := range rawItems {
		itemIDMap[raw.ID] = ItemKind(i)
		clr := hexToRGBA(raw.Color)
		var et EquipType
		var atk, def int
		if raw.Effect != nil && raw.Effect.Type == "equip" {
			switch raw.Effect.EquipType {
			case "weapon":
				et = EquipWeapon
			case "shield":
				et = EquipShield
			case "armor":
				et = EquipArmor
			}
			atk = raw.Effect.Atk
			def = raw.Effect.Def
		}
		itemDefs[i] = itemDef{
			name:       raw.Name,
			desc:       raw.Desc,
			effectDesc: raw.EffectDesc,
			weight:     raw.Weight,
			rarity:     raw.Rarity,
			clr:        clr,
			floorBound: raw.FloorBound,
			canUse:     buildCondition(raw.CanUse),
			effect:     buildEffect(raw.Effect),
			equipType:  et,
			atk:        atk,
			def:        def,
		}
	}
}

func buildEffect(spec *parameterizedEffectSpec) ItemEffect {
	if spec == nil {
		return nil
	}
	switch spec.Type {
	case "heal":
		return func(g *GameScene, _ int) {
			g.playerHP += spec.Amount
			if g.playerHP > playerMaxHP {
				g.playerHP = playerMaxHP
			}
			g.pushMessage(spec.Message)
		}
	case "reveal_map":
		return func(g *GameScene, _ int) {
			for x := 0; x < mapWidth; x++ {
				for y := 0; y < mapHeight; y++ {
					if g.worldMap[x][y] != Wall {
						g.explored[x][y] = true
					}
				}
			}
			if strings.Contains(spec.Message, "%d") {
				g.pushMessage(fmt.Sprintf(spec.Message, g.floor+1))
			} else {
				g.pushMessage(spec.Message)
			}
		}
	case "fireball":
		return func(g *GameScene, idx int) {
			if g.MP < spec.Cost {
				g.pushMessage("MPが足りない！")
				return
			}
			dx, dy := g.playerDir.delta()
			nx, ny := g.playerX+dx, g.playerY+dy
			
			hit := false
			for i := range g.enemies {
				e := &g.enemies[i]
				if e.x == nx && e.y == ny {
					g.MP -= spec.Cost
					g.pushMessage(spec.Message)
					e.hp -= spec.Damage
					if e.hp <= 0 {
						g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
						g.pushMessage(fmt.Sprintf("%sを倒した！", enemyDefs[e.kind].name))
					} else {
						g.pushMessage(fmt.Sprintf("%sに%dのダメージ！", enemyDefs[e.kind].name, spec.Damage))
					}
					hit = true
					break
				}
			}
			if !hit {
				g.pushMessage("何もない方向に炎を放った。")
			}
		}
	case "equip":
		return func(g *GameScene, idx int) {
			entry := &g.inventory[idx]
			def := &itemDefs[entry.kind]

			if entry.Equipped {
				entry.Equipped = false
				g.pushMessage(def.name + "を外した。")
				return
			}

			// 同じ箇所の他の装備を外す
			for i := range g.inventory {
				if g.inventory[i].Equipped && itemDefs[g.inventory[i].kind].equipType == def.equipType {
					g.inventory[i].Equipped = false
				}
			}
			entry.Equipped = true
			g.pushMessage(def.name + "を装備した。")
		}
	default:
		return nil
	}
}

func buildCondition(spec *parameterizedConditionSpec) func(*GameScene, InventoryEntry) (bool, string) {
	if spec == nil {
		return nil
	}
	switch spec.Type {
	case "check_same_floor":
		return func(g *GameScene, entry InventoryEntry) (bool, string) {
			if entry.obtainedSeed != g.mapSeed {
				return false, fmt.Sprintf(spec.Message, entry.obtainedFloor+1)
			}
			return true, ""
		}
	default:
		return nil
	}
}

func hexToRGBA(hex string) color.RGBA {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return color.RGBA{255, 255, 255, 255}
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

// MapItem はマップ上に落ちているアイテム
type MapItem struct {
	x, y          int
	kind          ItemKind
	obtainedSeed  int64 // floorBound アイテム用のマップ生成シード（それ以外は 0）
	obtainedFloor int   // ラベル表示用フロア番号（floorBound アイテムのみ有効）
}

// InventoryEntry はインベントリの1エントリ
type InventoryEntry struct {
	kind          ItemKind
	count         int
	obtainedSeed  int64 // floorBound アイテム用のマップ生成シード（それ以外は 0）
	obtainedFloor int   // ラベル表示用フロア番号（floorBound アイテムのみ有効）
	Equipped      bool
}

// currentWeight は現在の所持重量を返す
func (g *GameScene) currentWeight() int {
	total := 0
	for _, e := range g.inventory {
		total += itemDefs[e.kind].weight * e.count
	}
	return total
}

func (g *GameScene) equippedAtk() int {
	atk := 0
	for _, e := range g.inventory {
		if e.Equipped {
			atk += itemDefs[e.kind].atk
		}
	}
	return atk
}

func (g *GameScene) equippedDef() int {
	def := 0
	for _, e := range g.inventory {
		if e.Equipped {
			def += itemDefs[e.kind].def
		}
	}
	return def
}

// addInventory は重量上限を確認してからインベントリにアイテムを追加する。
// obtainedSeed は floorBound アイテムのスタック分割に使用する。
func (g *GameScene) addInventory(k ItemKind, obtainedSeed int64, obtainedFloor int) bool {
	def := &itemDefs[k]
	seed := int64(0)
	floor := 0
	if def.floorBound {
		seed = obtainedSeed
		floor = obtainedFloor
	}
	if g.currentWeight()+def.weight > maxCarryWeight {
		return false
	}
	// 装備品はスタックしない（Equipped 状態の管理のため）
	if def.equipType == EquipNone {
		for i := range g.inventory {
			if g.inventory[i].kind == k && g.inventory[i].obtainedSeed == seed {
				g.inventory[i].count++
				return true
			}
		}
	}
	g.inventory = append(g.inventory, InventoryEntry{kind: k, count: 1, obtainedSeed: seed, obtainedFloor: floor})
	return true
}

// useInventoryItem はインベントリの指定インデックスのアイテムを1個使用する。
// canUse チェックに失敗した場合は g.message にメッセージを設定して false を返す。
func (g *GameScene) useInventoryItem(idx int) bool {
	if idx < 0 || idx >= len(g.inventory) {
		return false
	}
	entry := g.inventory[idx]
	def := &itemDefs[entry.kind]

	if def.canUse != nil {
		if ok, msg := def.canUse(g, entry); !ok {
			g.pushMessage(msg)
			return false
		}
	}

	if def.effect != nil {
		def.effect(g, idx)
	}

	// 装備品は使用しても消費されない
	if def.equipType != EquipNone {
		return true
	}

	g.inventory[idx].count--
	if g.inventory[idx].count <= 0 {
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	}
	return true
}

// dropInventoryItem はインベントリの指定インデックスのアイテムを1個足元に捨てる。
func (g *GameScene) dropInventoryItem(idx int) {
	if idx < 0 || idx >= len(g.inventory) {
		return
	}

	// 足元に既にアイテムがないか確認
	_, alreadyOnFloor := g.itemAtFeet()
	if alreadyOnFloor {
		g.pushMessage("足元にアイテムが落ちているため、捨てられない。")
		return
	}

	entry := g.inventory[idx]
	def := &itemDefs[entry.kind]

	g.mapItems = append(g.mapItems, MapItem{
		x:             g.playerX,
		y:             g.playerY,
		kind:          entry.kind,
		obtainedSeed:  entry.obtainedSeed,
		obtainedFloor: entry.obtainedFloor,
	})

	g.pushMessage(def.name + "を捨てた。")
	g.inventory[idx].count--
	if g.inventory[idx].count <= 0 {
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	}
}

// pickupItem はプレイヤー位置のアイテムを取得する（重量上限チェックあり）
func (g *GameScene) pickupItem(idx int) {
	it := g.mapItems[idx]
	if !g.addInventory(it.kind, it.obtainedSeed, it.obtainedFloor) {
		g.pushMessage("重すぎて持てない！（重量: " + itoa(g.currentWeight()) + "/" + itoa(maxCarryWeight) + "）")
		return
	}
	g.mapItems = append(g.mapItems[:idx], g.mapItems[idx+1:]...)
	g.pushMessage(itemDefs[it.kind].name + "を手に入れた！")
}

// itoa は非負整数を文字列に変換する
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [10]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
