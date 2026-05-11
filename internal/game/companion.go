package game

import (
	"math/rand"
	"time"
)

type CompanionOrder int

const (
	OrderFollow CompanionOrder = iota
	OrderWait
	OrderAggressive
)

var CompanionOrderNames = map[CompanionOrder]string{
	OrderFollow:     "ついてきて",
	OrderWait:       "待ってて",
	OrderAggressive: "積極的に戦って",
}

type Companion struct {
	Actor
	kind  int
	order CompanionOrder
}

func (c *Companion) GetName() string {
	return companionDefs[c.kind].Name
}

func (c *Companion) GetRewardXP() int {
	return companionDefs[c.kind].XP
}

func (c *Companion) GetRelation(target Battler) RelationState {
	// プレイヤーとは常に友好
	if target.GetID() == 1 { // Player ID
		return RelationFriendly
	}
	// 他の仲間とも友好
	// TODO: ID範囲で判定するか、タイプを持たせる

	// 基本は Actor の Relation を見るが、
	// エネミーに対してはデフォルト敵対
	score := c.Relations[target.GetID()]
	def := &companionDefs[c.kind]

	if score <= def.NeutralThreshold {
		return RelationHostile
	}
	if score >= def.FriendlyThreshold {
		return RelationFriendly
	}

	// オーダーが Aggressive なら中立も敵対とみなす
	if c.order == OrderAggressive {
		return RelationHostile
	}

	return RelationNeutral
}

func (g *GameScene) moveSingleCompanion(idx int) {
	if idx >= len(g.companions) {
		return
	}
	c := &g.companions[idx]

	// 視界内の敵対ユニットを探す
	var target Battler
	// 敵を探す
	for i := range g.enemies {
		e := &g.enemies[i]
		if g.canSeeUnitCompanion(c, e.X, e.Y) {
			rel := c.GetRelation(e)
			if rel == RelationHostile {
				target = e
				break
			}
		}
	}

	if target != nil {
		// 攻撃範囲チェック (隣接)
		tx, ty := 0, 0
		// ターゲットの位置を取得
		// 汎用的に Battler から位置を取れるようにするか、ここで検索する
		for _, e := range g.enemies {
			if e.ID == target.GetID() {
				tx, ty = e.X, e.Y
				break
			}
		}

		dx := tx - c.X
		dy := ty - c.Y
		if abs(dx)+abs(dy) == 1 {
			c.Dir = dirFromDelta(dx, dy)
			c.AttackAnim = attackAnimFrames
			def := &companionDefs[c.kind]
			if def.FriendlyFire && g.companionFriendlyFireCheck(c, tx, ty) {
				return
			}
			// 敵に攻撃
			var targetEnemy *Enemy
			for i := range g.enemies {
				if g.enemies[i].ID == target.GetID() {
					targetEnemy = &g.enemies[i]
					break
				}
			}
			if targetEnemy != nil {
				g.Combat.ResolveCombat(g.Bus, c, targetEnemy, c.GetCombatType(), 0, c.GetStats().Element)
			}
			return
		}

		// 移動 (待機中でなければ)
		if c.order != OrderWait {
			nx, ny := g.bfsNextStep(c.X, c.Y, tx, ty)
			if nx != c.X || ny != c.Y {
				c.Dir = dirFromDelta(nx-c.X, ny-c.Y)
				if !g.isUnitAt(nx, ny) {
					c.X, c.Y = nx, ny
				}
			}
		}
	} else {
		// 敵がいない場合
		if c.order == OrderFollow || c.order == OrderAggressive {
			// プレイヤーに追従
			dist := abs(c.X-g.Player.X) + abs(c.Y-g.Player.Y)
			if dist > 1 {
				nx, ny := g.bfsNextStep(c.X, c.Y, g.Player.X, g.Player.Y)
				if nx != c.X || ny != c.Y {
					c.Dir = dirFromDelta(nx-c.X, ny-c.Y)
					if !g.isUnitAt(nx, ny) {
						c.X, c.Y = nx, ny
					}
				}
			} else {
				// 隣接しているなら、プレイヤーの方を向く
				c.Dir = dirFromDelta(g.Player.X-c.X, g.Player.Y-c.Y)
			}
		} else {
			// 待機中ならランダム移動or何もしない
			// 今回は「待機」なので動かないことにする
		}
	}
}

// フレンドリーファイア判定: 攻撃時に味方を巻き込む可能性をチェック
// true を返した場合、フレンドリーファイアが発生して攻撃がそちらに向いた
func (g *GameScene) companionFriendlyFireCheck(c *Companion, targetX, targetY int) bool {
	// 攻撃方向の左右に味方がいるかチェック（隣接マス）
	dx := targetX - c.X
	dy := targetY - c.Y

	// 攻撃先を含む隣接マスを確認
	var adjacentFriendlies []Battler
	dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	for _, d := range dirs {
		ax, ay := targetX+d[0], targetY+d[1]
		if ax == c.X && ay == c.Y {
			continue
		}
		if g.Player.X == ax && g.Player.Y == ay {
			adjacentFriendlies = append(adjacentFriendlies, g.Player)
		}
		for i := range g.companions {
			if g.companions[i].ID == c.ID {
				continue
			}
			if g.companions[i].X == ax && g.companions[i].Y == ay {
				adjacentFriendlies = append(adjacentFriendlies, &g.companions[i])
			}
		}
	}

	if len(adjacentFriendlies) == 0 {
		return false
	}

	// 仲間の攻撃が逸れて味方に当たる確率: 15%
	if rand.Intn(100) < 15 {
		victim := adjacentFriendlies[rand.Intn(len(adjacentFriendlies))]
		c.Dir = dirFromDelta(dx, dy)
		g.Combat.ResolveCombat(g.Bus, c, victim, c.GetCombatType(), 0, c.GetStats().Element)
		return true
	}
	return false
}

func (g *GameScene) canSeeUnitCompanion(c *Companion, targetX, targetY int) bool {
	// Enemy の canSeeUnit と同様のロジック
	// TODO: 共通化
	er := g.roomOf(c.X, c.Y)
	tr := g.roomOf(targetX, targetY)
	if er != -1 && er == tr {
		return true
	}
	if c.X == targetX {
		minY, maxY := c.Y, targetY
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY + 1; y < maxY; y++ {
			if g.worldMap[c.X][y] == Wall {
				return false
			}
		}
		return true
	}
	if c.Y == targetY {
		minX, maxX := c.X, targetX
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX + 1; x < maxX; x++ {
			if g.worldMap[x][c.Y] == Wall {
				return false
			}
		}
		return true
	}
	return false
}

func (g *GameScene) RemoveCompanion(ID int64) bool {
	for idx, c := range g.companions {
		if c.ID == ID {
			g.companions = append(g.companions[:idx], g.companions[idx+1:]...)
			return true
		}
	}
	return false
}

func (g *GameScene) isCompanionAt(x, y int) bool {
	for _, c := range g.companions {
		if c.X == x && c.Y == y {
			return true
		}
	}
	return false
}

func (g *GameScene) isUnitAt(x, y int) bool {
	if g.Player.X == x && g.Player.Y == y {
		return true
	}
	for _, e := range g.enemies {
		if e.X == x && e.Y == y {
			return true
		}
	}
	for _, c := range g.companions {
		if c.X == x && c.Y == y {
			return true
		}
	}
	return false
}

func (g *GameScene) spawnCompanion(kindID string, x, y int) {
	idx, ok := companionIDMap[kindID]
	if !ok {
		return
	}
	def := companionDefs[idx]
	c := Companion{
		Actor: Actor{
			ID:        time.Now().UnixNano(), // 適当なID
			X:         x,
			Y:         y,
			HP:        def.HP,
			MaxHP:     def.HP,
			MP:        def.MP,
			MaxMP:     def.MP,
			Str:       def.Str,
			Wis:       def.Wis,
			Fai:       def.Fai,
			Vit:       def.Vit,
			Agi:       def.Agi,
			Luk:       def.Luk,
			Element:   def.Element,
			Race:      def.Race,
			Relations: make(map[int64]int),
		},
		kind:  idx,
		order: OrderFollow,
	}
	g.companions = append(g.companions, c)
}
