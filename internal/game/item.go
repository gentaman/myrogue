package game

import (
	"fmt"
	"image/color"
)

const (
	playerMaxHP    = 20
	maxCarryWeight = 10
)

// ItemKind はアイテムの種別を表す
type ItemKind int

const (
	ItemHealSmall ItemKind = iota
	ItemHealLarge
	ItemRevealMap   // 取得したフロアでのみ使用可
	ItemClairvoyant // どのフロアでも使用可、使用フロアのみ全探索
	itemKindCount
)

// ItemEffect はアイテム使用時の効果コマンド
type ItemEffect func(g *GameScene, entry InventoryEntry)

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
}

// entryName はインベントリ行に表示する名前を返す（floorBound アイテムはフロア番号を付加）
func (d *itemDef) entryName(entry InventoryEntry) string {
	if d.floorBound {
		return fmt.Sprintf("%s[F%d]", d.name, entry.obtainedFloor+1)
	}
	return d.name
}

var itemDefs = []itemDef{
	ItemHealSmall: {
		name:       "回復薬（小）",
		desc:       "草を煎じた素朴な薬。飲むとHPが5回復する。軽くて持ち運びやすい。",
		effectDesc: "HP +5",
		weight:     2,
		rarity:     1,
		clr:        color.RGBA{100, 255, 150, 255},
		effect: func(g *GameScene, _ InventoryEntry) {
			g.playerHP += 5
			if g.playerHP > playerMaxHP {
				g.playerHP = playerMaxHP
			}
			g.message = "回復薬（小）を使った！ HP +5"
		},
	},
	ItemHealLarge: {
		name:       "回復薬（大）",
		desc:       "希少な霊草から作られた上質な薬。飲むとHPが10回復する。効果は高いが重い。",
		effectDesc: "HP +10",
		weight:     4,
		rarity:     2,
		clr:        color.RGBA{50, 220, 80, 255},
		effect: func(g *GameScene, _ InventoryEntry) {
			g.playerHP += 10
			if g.playerHP > playerMaxHP {
				g.playerHP = playerMaxHP
			}
			g.message = "回復薬（大）を使った！ HP +10"
		},
	},
	ItemRevealMap: {
		name:       "フロアの地図",
		desc:       "特定のフロアの全体像が描かれた古い地図。取得したフロアでのみ使用できる。他のフロアに持ち込んでも使えない。",
		effectDesc: "取得フロアのみ全探索",
		weight:     1,
		rarity:     1,
		floorBound: true,
		clr:        color.RGBA{255, 220, 80, 255},
		canUse: func(g *GameScene, entry InventoryEntry) (bool, string) {
			if entry.obtainedSeed != g.mapSeed {
				return false, fmt.Sprintf("この地図はフロア%dでしか使えない。", entry.obtainedFloor+1)
			}
			return true, ""
		},
		effect: func(g *GameScene, _ InventoryEntry) {
			for x := 0; x < mapWidth; x++ {
				for y := 0; y < mapHeight; y++ {
					if g.worldMap[x][y] != Wall {
						g.explored[x][y] = true
					}
				}
			}
			g.message = "フロアの地図を使った！このフロアの全体が明らかになった。"
		},
	},
	ItemClairvoyant: {
		name:       "千里眼の薬",
		desc:       "遠く離れた場所を見通す不思議な薬。使ったフロアの全体が明らかになる。どのフロアでも使えるので、深い階層まで温存するのも手だ。",
		effectDesc: "使用フロアのみ全探索",
		weight:     2,
		rarity:     3,
		clr:        color.RGBA{100, 180, 255, 255},
		effect: func(g *GameScene, _ InventoryEntry) {
			for x := 0; x < mapWidth; x++ {
				for y := 0; y < mapHeight; y++ {
					if g.worldMap[x][y] != Wall {
						g.explored[x][y] = true
					}
				}
			}
			g.message = fmt.Sprintf("千里眼の薬を飲んだ！フロア%dの全体が明らかになった。", g.floor+1)
		},
	},
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
}

// currentWeight は現在の所持重量を返す
func (g *GameScene) currentWeight() int {
	total := 0
	for _, e := range g.inventory {
		total += itemDefs[e.kind].weight * e.count
	}
	return total
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
	for i := range g.inventory {
		if g.inventory[i].kind == k && g.inventory[i].obtainedSeed == seed {
			g.inventory[i].count++
			return true
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
			g.message = msg
			return false
		}
	}

	def.effect(g, entry)

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
	entry := g.inventory[idx]
	def := &itemDefs[entry.kind]

	alreadyOnFloor := false
	for _, it := range g.mapItems {
		if it.x == g.playerX && it.y == g.playerY {
			alreadyOnFloor = true
			break
		}
	}
	if !alreadyOnFloor {
		g.mapItems = append(g.mapItems, MapItem{
			x:             g.playerX,
			y:             g.playerY,
			kind:          entry.kind,
			obtainedSeed:  entry.obtainedSeed,
			obtainedFloor: entry.obtainedFloor,
		})
	}

	g.message = def.name + "を捨てた。"
	g.inventory[idx].count--
	if g.inventory[idx].count <= 0 {
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	}
}

// pickupItem はプレイヤー位置のアイテムを取得する（重量上限チェックあり）
func (g *GameScene) pickupItem() {
	for i, it := range g.mapItems {
		if it.x == g.playerX && it.y == g.playerY {
			if !g.addInventory(it.kind, it.obtainedSeed, it.obtainedFloor) {
				g.message = "重すぎて持てない！（重量: " + itoa(g.currentWeight()) + "/" + itoa(maxCarryWeight) + "）"
				return
			}
			g.mapItems = append(g.mapItems[:i], g.mapItems[i+1:]...)
			g.message = itemDefs[it.kind].name + "を手に入れた！"
			return
		}
	}
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
