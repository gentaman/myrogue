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

type ItemEffect func(g *GameScene, actor *Actor, entryIdx int)

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
	durability int
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
	// 耐久度を表示 (消耗しないアイテム以外)
	if d.durability > 0 {
		name = fmt.Sprintf("%s (%d)", name, entry.Durability)
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
	Durability int                         `json:"durability"`
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
			durability: raw.Durability,
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
		spec = &parameterizedEffectSpec{Type: "nop"}
	}

	switch spec.Type {
	case "heal":
		return func(g *GameScene, actor *Actor, _ int) {
			actor.HP += spec.Amount
			if actor.HP > actor.MaxHP {
				actor.HP = actor.MaxHP
			}
			g.pushMessage(spec.Message)
		}
	case "heal_faith":
		return func(g *GameScene, actor *Actor, _ int) {
			heal := 5 + actor.Fai*2
			actor.HP += heal
			if actor.HP > actor.MaxHP {
				actor.HP = actor.MaxHP
			}
			g.pushMessage(spec.Message)
		}
	case "reveal_map":
		return func(g *GameScene, _ *Actor, _ int) {
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
		return func(g *GameScene, actor *Actor, idx int) {
			if actor.MP < spec.Cost {
				if actor.ID == g.Player.ID {
					g.pushMessage("MPが足りない！")
				}
				return
			}
			dx, dy := actor.Dir.delta()
			nx, ny := actor.X+dx, actor.Y+dy
			hit := false
			target := g.unitAt(nx, ny)
			if target != nil {
				actor.MP -= spec.Cost
				if actor.ID == g.Player.ID {
					g.pushMessage(spec.Message)
				}
				target.ApplyDamage(spec.Damage)
				target.UpdateRelation(actor.GetID(), -spec.Damage*10)

				// ユニットが死亡したかチェック
				if target.GetID() == g.Player.ID {
					if g.Player.HP <= 0 {
						g.playState = StateDead
						g.pushMessage("あなたは力尽きた...")
					}
				} else {
					for i := range g.enemies {
						e := &g.enemies[i]
						if e.ID == target.GetID() && e.HP <= 0 {
							g.pushMessage(fmt.Sprintf("%sを倒した！", e.GetName()))
							g.actorGainXP(actor, enemyDefs[e.kind].xp)
							g.dropChest(e.X, e.Y, e.Inventory)
							g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
							break
						}
					}
				}
				hit = true
			}
			if !hit && actor.ID == g.Player.ID {
				g.pushMessage("何もない方向に炎を放った。")
			}
		}
	case "ranged_magic":
		return func(g *GameScene, actor *Actor, idx int) {
			if actor.MP < spec.Cost {
				if actor.ID == g.Player.ID {
					g.pushMessage("MPが足りない！")
				}
				return
			}
			actor.MP -= spec.Cost
			dx, dy := actor.Dir.delta()
			targetX, targetY := actor.X, actor.Y
			var targetUnit Battler
			for r := 1; r <= spec.Range; r++ {
				tx, ty := actor.X+dx*r, actor.Y+dy*r
				if tx < 0 || tx >= mapWidth || ty < 0 || ty >= mapHeight || g.worldMap[tx][ty] == Wall {
					break
				}
				targetX, targetY = tx, ty
				targetUnit = g.unitAt(tx, ty)
				if targetUnit != nil {
					break
				}
			}
			g.projectiles = append(g.projectiles, Projectile{
				StartX:      float64(actor.X*tileSize + tileSize/2),
				StartY:      float64(actor.Y*tileSize + tileSize/2),
				EndX:        float64(targetX*tileSize + tileSize/2),
				EndY:        float64(targetY*tileSize + tileSize/2),
				Frame:       0,
				TotalFrames: 15,
				Color:       hexToRGBA(spec.Color),
			})
			if actor.ID == g.Player.ID {
				g.pushMessage(spec.Message)
			}
			if targetUnit != nil {
				targetUnit.ApplyDamage(spec.Damage)
				targetUnit.UpdateRelation(actor.GetID(), -spec.Damage*10)
				if targetUnit.GetID() == g.Player.ID {
					if g.Player.HP <= 0 {
						g.playState = StateDead
						g.pushMessage("あなたは力尽きた...")
					}
				} else {
					for i := range g.enemies {
						e := &g.enemies[i]
						if e.ID == targetUnit.GetID() && e.HP <= 0 {
							g.pushMessage(fmt.Sprintf("%sを倒した！", e.GetName()))
							g.actorGainXP(actor, enemyDefs[e.kind].xp)
							g.dropChest(e.X, e.Y, e.Inventory)
							g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
							break
						}
					}
				}
			}
		}
	case "equip":
		return func(g *GameScene, actor *Actor, idx int) {
			entry := &actor.Inventory[idx]
			def := &itemDefs[entry.kind]

			if entry.Equipped {
				entry.Equipped = false
				if actor.ID == g.Player.ID {
					g.pushMessage(def.name + "を外した。")
				}
				return
			}

			// 同じ箇所の他の装備を外す
			for i := range actor.Inventory {
				if actor.Inventory[i].Equipped && itemDefs[actor.Inventory[i].kind].equipType == def.equipType {
					actor.Inventory[i].Equipped = false
				}
			}
			entry.Equipped = true
			if actor.ID == g.Player.ID {
				g.pushMessage(def.name + "を装備した。")
			}
		}
	default:
		return func(g *GameScene, actor *Actor, idx int) {
			entry := &actor.Inventory[idx]
			def := &itemDefs[entry.kind]
			if actor.ID == g.Player.ID {
				g.pushMessage(def.name + "を使ったが、何もおきなかった。")
			}
		}
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
	X, Y      int
	Inventory []InventoryEntry
}

type InventoryEntry struct {
	kind          ItemKind
	count         int
	obtainedSeed  int64
	obtainedFloor int
	Equipped      bool
	Durability    int
}

func (a *Actor) currentWeight() int {
	total := 0
	for _, e := range a.Inventory {
		total += itemDefs[e.kind].weight * e.count
	}
	return total
}

func (a *Actor) equippedAtk() int {
	atk := 0
	for _, e := range a.Inventory {
		if e.Equipped {
			atk += itemDefs[e.kind].atk
		}
	}
	return atk
}

func (a *Actor) equippedDef() int {
	def := 0
	for _, e := range a.Inventory {
		if e.Equipped {
			def += itemDefs[e.kind].def
		}
	}
	return def
}

func (g *GameScene) currentWeight() int {
	return g.Player.currentWeight()
}

func (g *GameScene) equippedAtk() int {
	return g.Player.equippedAtk()
}

func (g *GameScene) equippedDef() int {
	return g.Player.equippedDef()
}

func (g *GameScene) addInventory(actor *Actor, k ItemKind, durability int, obtainedSeed int64, obtainedFloor int) bool {
	def := &itemDefs[k]
	seed, floor := int64(0), 0
	if def.floorBound {
		seed, floor = obtainedSeed, obtainedFloor
	}

	// プレイヤーのみ重量チェック
	if actor.ID == g.Player.ID {
		if g.currentWeight()+def.weight > maxCarryWeight {
			return false
		}
	}

	// 無限耐久度の消耗品（あれば）のみスタック
	if def.durability < 0 && def.equipType == EquipNone {
		for i := range actor.Inventory {
			if actor.Inventory[i].kind == k && actor.Inventory[i].obtainedSeed == seed {
				actor.Inventory[i].count++
				return true
			}
		}
	}

	actor.Inventory = append(actor.Inventory, InventoryEntry{
		kind:          k,
		count:         1,
		obtainedSeed:  seed,
		obtainedFloor: floor,
		Durability:    durability,
	})
	return true
}

func (g *GameScene) useInventoryItem(idx int) bool {
	if idx < 0 || idx >= len(g.Player.Inventory) {
		return false
	}
	entry := &g.Player.Inventory[idx]
	def := &itemDefs[entry.kind]
	if def.canUse != nil {
		if ok, msg := def.canUse(g, *entry); !ok {
			g.pushMessage(msg)
			return false
		}
	}
	if def.effect != nil {
		def.effect(g, &g.Player, idx)
	}

	// 装備品の使用（着脱）では耐久度を減らさない（攻撃/防御時に減らす）
	if def.equipType != EquipNone {
		return true
	}

	// 消耗品の使用
	if entry.Durability > 0 {
		entry.Durability--
		if entry.Durability <= 0 {
			g.pushMessage(def.name + "は壊れた。")
			g.Player.Inventory = append(g.Player.Inventory[:idx], g.Player.Inventory[idx+1:]...)
		}
	}
	return true
}

func (g *GameScene) dropInventoryItem(idx int) {
	if idx < 0 || idx >= len(g.Player.Inventory) {
		return
	}

	entry := g.Player.Inventory[idx]
	// 装備中は外す
	if entry.Equipped {
		entry.Equipped = false
	}

	g.dropChest(g.Player.X, g.Player.Y, []InventoryEntry{entry})

	g.pushMessage(itemDefs[entry.kind].name + "を捨てた。")
	g.Player.Inventory = append(g.Player.Inventory[:idx], g.Player.Inventory[idx+1:]...)
}

func (g *GameScene) pickupItem(idx int) {
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
