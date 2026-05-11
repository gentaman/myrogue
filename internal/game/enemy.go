package game

import (
	"math/rand"
)

// Dir はキャラクターの向きを表す
type Dir int

const (
	DirUp Dir = iota
	DirDown
	DirLeft
	DirRight
)

func (d Dir) delta() (int, int) {
	switch d {
	case DirUp:
		return 0, -1
	case DirDown:
		return 0, 1
	case DirLeft:
		return -1, 0
	default:
		return 1, 0
	}
}

// EnemyState は敵の警戒状態を表す
type EnemyState int

const (
	EnemyIdle    EnemyState = iota // プレイヤーに気づいていない
	EnemyAlerted                   // プレイヤーを発見・追跡中
)

// EnemyKind は敵の種別を表す
type EnemyKind int

// Enemy は敵キャラクターを表す
type Enemy struct {
	Actor
	state    EnemyState // 警戒状態
	kind     EnemyKind
	rewardXP int
}

func (e *Enemy) GetName() string { return enemyDefs[e.kind].Name }

func (e *Enemy) GetRelation(target Battler) RelationState {
	def := &enemyDefs[e.kind]
	// 同じ種族なら常に友好
	if e.Race == target.GetRace() {
		return RelationFriendly
	}
	score := e.Relations[target.GetID()]
	if score <= def.NeutralThreshold {
		return RelationHostile
	}
	if score >= def.FriendlyThreshold {
		return RelationFriendly
	}
	// デフォルトは種族によって変えても良いが、一旦は敵対（または中立）
	// ローグライクの敵なので、プレイヤーや他種族にはデフォルト敵対 (-100) で初期化することにする
	return RelationHostile
}

func (e *Enemy) ConsumeDurability(bus *EventBus, et EquipType) {
	if et == EquipNone {
		return
	}
	for i := 0; i < len(e.Inventory); i++ {
		entry := &e.Inventory[i]
		if entry.Equipped && itemDefs[entry.kind].equipType == et {
			def := &itemDefs[entry.kind]
			if def.durability < 0 {
				return
			}
			entry.Durability--
			if entry.Durability <= 0 {
				// 敵の装備が壊れた時のメッセージは必要なら追加
				e.Inventory = append(e.Inventory[:i], e.Inventory[i+1:]...)
			}
			return
		}
	}
}

func (e *Enemy) GetRewardXP() int {
	return e.rewardXP
}

func (g *GameScene) isEnemyAt(x, y int) bool {
	for _, e := range g.enemies {
		if e.X == x && e.Y == y {
			return true
		}
	}
	return false
}

// 特定のマスにいるユニット(敵 or プレイヤー or 仲間)を返す
func (g *GameScene) unitAt(x, y int) Battler {
	if g.Player.X == x && g.Player.Y == y {
		return g.Player
	}
	for i := range g.enemies {
		if g.enemies[i].X == x && g.enemies[i].Y == y {
			return &g.enemies[i]
		}
	}
	for i := range g.companions {
		if g.companions[i].X == x && g.companions[i].Y == y {
			return &g.companions[i]
		}
	}
	return nil
}

func (g *GameScene) attackEnemy(x, y int) {
	for i := range g.enemies {
		e := &g.enemies[i]
		if e.X == x && e.Y == y {
			// g.Combat.AttackEnemy(g, i)
			g.Combat.ResolveCombat(g.Bus, g.Player, e, g.Player.GetCombatType(), 0, g.Player.GetStats().Element)
			g.Bus.Publish(MsgSFX{PCM: sfxHitPCM})
			return
		}
	}
}

func dirFromDelta(dx, dy int) Dir {
	switch {
	case dx > 0:
		return DirRight
	case dx < 0:
		return DirLeft
	case dy > 0:
		return DirDown
	default:
		return DirUp
	}
}

func (g *GameScene) canSeeUnit(e *Enemy, targetX, targetY int) bool {
	er := g.roomOf(e.X, e.Y)
	tr := g.roomOf(targetX, targetY)
	if er != -1 && er == tr {
		return true
	}
	if e.X == targetX {
		minY, maxY := e.Y, targetY
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY + 1; y < maxY; y++ {
			if g.worldMap[e.X][y] == Wall {
				return false
			}
		}
		return true
	}
	if e.Y == targetY {
		minX, maxX := e.X, targetX
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX + 1; x < maxX; x++ {
			if g.worldMap[x][e.Y] == Wall {
				return false
			}
		}
		return true
	}
	return false
}

func (g *GameScene) bfsNextStep(startX, startY, goalX, goalY int) (int, int) {
	if startX == goalX && startY == goalY {
		return startX, startY
	}
	type pos struct{ x, y int }
	prev := [mapWidth][mapHeight]pos{}
	visited := [mapWidth][mapHeight]bool{}
	visited[startX][startY] = true
	queue := []pos{{startX, startY}}
	dirs := [4]pos{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	found := false
	for len(queue) > 0 && !found {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := cur.x+d.x, cur.y+d.y
			if nx < 0 || nx >= mapWidth || ny < 0 || ny >= mapHeight {
				continue
			}
			if visited[nx][ny] || g.worldMap[nx][ny] == Wall {
				continue
			}
			visited[nx][ny] = true
			prev[nx][ny] = cur
			if nx == goalX && ny == goalY {
				found = true
				break
			}
			queue = append(queue, pos{nx, ny})
		}
	}
	if !found {
		return startX, startY
	}
	cur := pos{goalX, goalY}
	for prev[cur.x][cur.y] != (pos{startX, startY}) {
		cur = prev[cur.x][cur.y]
	}
	return cur.x, cur.y
}

func (g *GameScene) bfsFleeStep(startX, startY, fearX, fearY int) (int, int) {
	type pos struct{ x, y int }
	dirs := [4]pos{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	bestX, bestY := startX, startY
	maxDist := float64((startX-fearX)*(startX-fearX) + (startY-fearY)*(startY-fearY))
	for _, d := range dirs {
		nx, ny := startX+d.x, startY+d.y
		if nx < 0 || nx >= mapWidth || ny < 0 || ny >= mapHeight || g.worldMap[nx][ny] == Wall || g.isEnemyAt(nx, ny) {
			continue
		}
		dist := float64((nx-fearX)*(nx-fearX) + (ny-fearY)*(ny-fearY))
		if dist > maxDist {
			maxDist = dist
			bestX, bestY = nx, ny
		}
	}
	return bestX, bestY
}

func (g *GameScene) moveSingleEnemy(idx int) {
	if idx >= len(g.enemies) {
		return
	}
	e := &g.enemies[idx]
	def := &enemyDefs[e.kind]

	// 視界内の敵対ユニットを探す（プレイヤー・仲間含む）
	var target Battler
	if g.canSeeUnit(e, g.Player.X, g.Player.Y) {
		if e.GetRelation(g.Player) == RelationHostile {
			target = g.Player
		}
	}
	if target == nil {
		for i := range g.companions {
			c := &g.companions[i]
			if g.canSeeUnit(e, c.X, c.Y) && e.GetRelation(c) == RelationHostile {
				target = c
				break
			}
		}
	}
	if target == nil {
		for i := range g.enemies {
			if i == idx {
				continue
			}
			en := &g.enemies[i]
			if g.canSeeUnit(e, en.X, en.Y) && e.GetRelation(en) == RelationHostile {
				target = en
				break
			}
		}
	}

	if target != nil {
		e.state = EnemyAlerted
		// 攻撃範囲チェック (隣接)
		tx, ty := 0, 0
		if target.GetID() == g.Player.ID {
			tx, ty = g.Player.X, g.Player.Y
		} else {
			found := false
			for _, c := range g.companions {
				if c.ID == target.GetID() {
					tx, ty = c.X, c.Y
					found = true
					break
				}
			}
			if !found {
				for _, en := range g.enemies {
					if en.ID == target.GetID() {
						tx, ty = en.X, en.Y
						break
					}
				}
			}
		}

		dx := tx - e.X
		dy := ty - e.Y
		if abs(dx)+abs(dy) == 1 {
			e.Dir = dirFromDelta(dx, dy)
			e.AttackAnim = attackAnimFrames
			if target.GetID() == g.Player.ID {
				g.Combat.ResolveCombat(g.Bus, e, g.Player, e.GetCombatType(), 0, e.GetStats().Element)
			} else {
				// 仲間への攻撃
				companionTarget := false
				for i := range g.companions {
					if g.companions[i].ID == target.GetID() {
						g.Combat.ResolveCombat(g.Bus, e, &g.companions[i], e.GetCombatType(), 0, e.GetStats().Element)
						companionTarget = true
						break
					}
				}
				if !companionTarget {
					// ユニット間戦闘（敵同士）
					for i := range g.enemies {
						if g.enemies[i].ID == target.GetID() {
							g.Combat.ResolveCombat(g.Bus, e, &g.enemies[i], e.GetCombatType(), 0, e.GetStats().Element)
							break
						}
					}
				}
			}
			return
		}

		// 移動
		var nx, ny int
		action := "pursue"
		switch def.Personality {
		case PersonalityCowardly:
			action = "flee"
		case PersonalityCalculated:
			if g.Player.Level > 1+g.floor*2 {
				action = "flee"
			}
		}

		if action == "flee" {
			nx, ny = g.bfsFleeStep(e.X, e.Y, tx, ty)
		} else {
			nx, ny = g.bfsNextStep(e.X, e.Y, tx, ty)
		}

		if nx != e.X || ny != e.Y {
			e.Dir = dirFromDelta(nx-e.X, ny-e.Y)
			if !g.isEnemyAt(nx, ny) && !(nx == g.Player.X && ny == g.Player.Y) && !g.isCompanionAt(nx, ny) {
				e.X, e.Y = nx, ny
			}
		}
	} else {
		e.state = EnemyIdle
		dirList := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
		candidates := dirList
		rand.Shuffle(len(candidates), func(a, b int) { candidates[a], candidates[b] = candidates[b], candidates[a] })
		nx, ny := e.X, e.Y
		for _, d := range candidates {
			cx, cy := e.X+d[0], e.Y+d[1]
			if cx >= 0 && cx < mapWidth && cy >= 0 && cy < mapHeight && g.worldMap[cx][cy] != Wall {
				nx, ny = cx, cy
				break
			}
		}
		if nx != e.X || ny != e.Y {
			e.Dir = dirFromDelta(nx-e.X, ny-e.Y)
			if !g.isEnemyAt(nx, ny) && !(nx == g.Player.X && ny == g.Player.Y) && !g.isCompanionAt(nx, ny) {
				e.X, e.Y = nx, ny
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
