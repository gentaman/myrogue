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

const (
	ItemHealSmall ItemKind = iota
	ItemHealLarge
	ItemRevealMap
	ItemClairvoyant
)

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
	rarity     int
	clr        color.RGBA
	floorBound bool
	canUse     func(g *GameScene, entry InventoryEntry) (bool, string)
	effect     ItemEffect

	equipType EquipType
	atk       int
	def       int
}

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

type parameterizedEffectSpec struct {
	Type      string `json:"type"`
	Amount    int    `json:"amount"`
	Message   string `json:"message"`
	EquipType string `json:"equip_type"`
	Atk       int    `json:"atk"`
	Def       int    `json:"def"`
	Cost      int    `json:"cost"`
	Damage    int    `json:"damage"`
	Range     int    `json:"range"`
	Color     string `json:"color"`
}

type parameterizedConditionSpec struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type rawItem struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Desc       string                      `json:"desc"`
	EffectDesc string                      `json:"effect_desc"`
	Weight     int                         `json:"weight"`
	Rarity     int                         `json:"rarity"`
	Color      string                      `json:"color"`
	FloorBound bool                        `json:"floor_bound"`
	CanUse     *parameterizedConditionSpec `json:"can_use"`
	Effect     *parameterizedEffectSpec    `json:"effect"`
}

func init() {
	var playerCfg struct {
		MaxHP          int `json:"max_hp"`
		MaxCarryWeight int `json:"max_carry_weight"`
	}
	if err := json.Unmarshal(playerJSON, &playerCfg); err != nil {
		panic(fmt.Sprintf("failed to unmarshal player.json: %v", err))
	}
	playerMaxHP = playerCfg.MaxHP
	maxCarryWeight = playerCfg.MaxCarryWeight

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
			g.Player.HP += spec.Amount
			if g.Player.HP > g.Player.MaxHP {
				g.Player.HP = g.Player.MaxHP
			}
			g.pushMessage(spec.Message)
		}
	case "heal_faith":
		return func(g *GameScene, _ int) {
			heal := 5 + g.Fai*2
			g.Player.HP += heal
			if g.Player.HP > g.Player.MaxHP {
				g.Player.HP = g.Player.MaxHP
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
			dx, dy := g.Player.Dir.delta()
			nx, ny := g.Player.X+dx, g.Player.Y+dy
			hit := false
			for i := range g.enemies {
				e := &g.enemies[i]
				if e.X == nx && e.Y == ny {
					g.MP -= spec.Cost
					g.pushMessage(spec.Message)
					e.HP -= spec.Damage
					e.DamageAnim = damageAnimFrames
					if e.HP <= 0 {
						g.pushMessage(fmt.Sprintf("%sを倒した！", enemyDefs[e.kind].name))
						g.gainXP(enemyDefs[e.kind].xp)
						g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
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
	case "ranged_magic":
		return func(g *GameScene, idx int) {
			if g.MP < spec.Cost {
				g.pushMessage("MPが足りない！")
				return
			}
			g.MP -= spec.Cost
			dx, dy := g.Player.Dir.delta()
			targetX, targetY := g.Player.X, g.Player.Y
			var targetEnemyIdx int = -1
			for r := 1; r <= spec.Range; r++ {
				tx, ty := g.Player.X+dx*r, g.Player.Y+dy*r
				if tx < 0 || tx >= mapWidth || ty < 0 || ty >= mapHeight || g.worldMap[tx][ty] == Wall {
					break
				}
				targetX, targetY = tx, ty
				found := false
				for i, e := range g.enemies {
					if e.X == tx && e.Y == ty {
						targetEnemyIdx = i
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			g.projectiles = append(g.projectiles, Projectile{
				StartX:      float64(g.Player.X*tileSize + tileSize/2),
				StartY:      float64(g.Player.Y*tileSize + tileSize/2),
				EndX:        float64(targetX*tileSize + tileSize/2),
				EndY:        float64(targetY*tileSize + tileSize/2),
				Frame:       0,
				TotalFrames: 15,
				Color:       hexToRGBA(spec.Color),
			})
			g.pushMessage(spec.Message)
			if targetEnemyIdx >= 0 {
				e := &g.enemies[targetEnemyIdx]
				def := &enemyDefs[e.kind]
				e.HP -= spec.Damage
				e.DamageAnim = damageAnimFrames
				if e.HP <= 0 {
					g.pushMessage(fmt.Sprintf("%sを倒した！", def.name))
					g.gainXP(def.xp)
					g.enemies = append(g.enemies[:targetEnemyIdx], g.enemies[targetEnemyIdx+1:]...)
				} else {
					g.pushMessage(fmt.Sprintf("%sに%dのダメージ！", def.name, spec.Damage))
				}
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

type MapItem struct {
	x, y          int
	kind          ItemKind
	obtainedSeed  int64
	obtainedFloor int
}

type InventoryEntry struct {
	kind          ItemKind
	count         int
	obtainedSeed  int64
	obtainedFloor int
	Equipped      bool
}

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

func (g *GameScene) addInventory(k ItemKind, obtainedSeed int64, obtainedFloor int) bool {
	def := &itemDefs[k]
	seed, floor := int64(0), 0
	if def.floorBound {
		seed, floor = obtainedSeed, obtainedFloor
	}
	if g.currentWeight()+def.weight > maxCarryWeight {
		return false
	}
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
	if def.equipType != EquipNone {
		return true
	}
	g.inventory[idx].count--
	if g.inventory[idx].count <= 0 {
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	}
	return true
}

func (g *GameScene) dropInventoryItem(idx int) {
	if idx < 0 || idx >= len(g.inventory) {
		return
	}
	if _, already := g.itemAtFeet(); already {
		g.pushMessage("足元にアイテムが落ちているため、捨てられない。")
		return
	}
	entry := g.inventory[idx]
	g.mapItems = append(g.mapItems, MapItem{
		x: g.Player.X, y: g.Player.Y,
		kind: entry.kind, obtainedSeed: entry.obtainedSeed, obtainedFloor: entry.obtainedFloor,
	})
	g.pushMessage(itemDefs[entry.kind].name + "を捨てた。")
	g.inventory[idx].count--
	if g.inventory[idx].count <= 0 {
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	}
}

func (g *GameScene) pickupItem(idx int) {
	it := g.mapItems[idx]
	if !g.addInventory(it.kind, it.obtainedSeed, it.obtainedFloor) {
		g.pushMessage("重すぎて持てない！（重量: " + itoa(g.currentWeight()) + "/" + itoa(maxCarryWeight) + "）")
		return
	}
	g.mapItems = append(g.mapItems[:idx], g.mapItems[idx+1:]...)
	g.pushMessage(itemDefs[it.kind].name + "を手に入れた！")
}

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
