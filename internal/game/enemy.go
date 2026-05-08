package game

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"
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

type enemyDef struct {
	name     string
	hp       int
	atk      int
	acc      int
	xp       int
	rarity   int
	clr      color.RGBA
	floorMin int
}

var (
	//go:embed assets/enemies.json
	enemiesJSON []byte

	enemyDefs []enemyDef

	enemyIDMap = map[string]EnemyKind{}
)

type rawEnemy struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	HP       int    `json:"hp"`
	ATK      int    `json:"atk"`
	ACC      int    `json:"acc"`
	XP       int    `json:"xp"`
	Rarity   int    `json:"rarity"`
	Color    string `json:"color"`
	FloorMin int    `json:"floor_min"`
}

func init() {
	var rawEnemies []rawEnemy
	if err := json.Unmarshal(enemiesJSON, &rawEnemies); err != nil {
		panic(fmt.Sprintf("failed to unmarshal enemies.json: %v", err))
	}

	enemyDefs = make([]enemyDef, len(rawEnemies))
	for i, raw := range rawEnemies {
		enemyIDMap[raw.ID] = EnemyKind(i)
		enemyDefs[i] = enemyDef{
			name:     raw.Name,
			hp:       raw.HP,
			atk:      raw.ATK,
			acc:      raw.ACC,
			xp:       raw.XP,
			rarity:   raw.Rarity,
			clr:      hexToRGBA(raw.Color),
			floorMin: raw.FloorMin,
		}
	}
}

// Enemy は敵キャラクターを表す
type Enemy struct {
	Actor
	state EnemyState // 警戒状態
	kind  EnemyKind
}

func (g *GameScene) isEnemyAt(x, y int) bool {
	for _, e := range g.enemies {
		if e.X == x && e.Y == y {
			return true
		}
	}
	return false
}

// プレイヤーが敵を攻撃
func (g *GameScene) attackEnemy(x, y int) {
	damage := 1 + g.Str + g.equippedAtk()
	for i := range g.enemies {
		e := &g.enemies[i]
		if e.X == x && e.Y == y {
			def := &enemyDefs[e.kind]
			e.HP -= damage
			e.DamageAnim = damageAnimFrames
			if e.HP <= 0 {
				g.pushMessage(fmt.Sprintf("%sを倒した！", def.name))
				g.gainXP(def.xp)
				g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
			} else {
				g.pushMessage(fmt.Sprintf("%sに %d のダメージを与えた！", def.name, damage))
			}
			playSFXHit()
			return
		}
	}
}

// 向き (dx,dy) を Dir に変換するヘルパー
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

// 敵がプレイヤーを視認できるか判定する
func (g *GameScene) canSeePlayer(e *Enemy) bool {
	er := g.roomOf(e.X, e.Y)
	pr := g.roomOf(g.Player.X, g.Player.Y)
	if er != -1 && er == pr {
		return true
	}
	if e.X == g.Player.X {
		minY, maxY := e.Y, g.Player.Y
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
	if e.Y == g.Player.Y {
		minX, maxX := e.X, g.Player.X
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

// BFS で (startX,startY) から (goalX,goalY) への次の1マスを返す
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

// 個別の敵の行動を処理
func (g *GameScene) moveSingleEnemy(idx int) {
	if idx >= len(g.enemies) {
		return
	}
	e := &g.enemies[idx]
	def := &enemyDefs[e.kind]

	// 正面にプレイヤーがいるなら攻撃
	fdx, fdy := e.Dir.delta()
	if e.X+fdx == g.Player.X && e.Y+fdy == g.Player.Y {
		e.AttackAnim = attackAnimFrames
		if g.tryDodge(def.acc) {
			g.pushMessage(fmt.Sprintf("%sの攻撃をかわした！", def.name))
		} else {
			damage := def.atk - (g.equippedDef() + g.Vit)
			if damage < 0 {
				damage = 0
			}
			g.Player.HP -= damage
			g.Player.DamageAnim = damageAnimFrames
			playSFXHit()
			if g.Player.HP <= 0 {
				g.Player.HP = 0
				g.playState = StateDead
				g.pushMessage("力尽きた...")
				return
			}
			if damage > 0 {
				g.pushMessage(fmt.Sprintf("%sに攻撃された！ %d のダメージ！ HP: %d", def.name, damage, g.Player.HP))
			} else {
				g.pushMessage(fmt.Sprintf("%sの攻撃を弾き返した！", def.name))
			}
		}
		return
	}

	// 視界チェック
	if g.canSeePlayer(e) {
		e.state = EnemyAlerted
	} else {
		e.state = EnemyIdle
	}

	var nx, ny int
	if e.state == EnemyAlerted {
		nx, ny = g.bfsNextStep(e.X, e.Y, g.Player.X, g.Player.Y)
	} else {
		dirList := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
		candidates := dirList
		rand.Shuffle(len(candidates), func(a, b int) { candidates[a], candidates[b] = candidates[b], candidates[a] })
		nx, ny = e.X, e.Y
		for _, d := range candidates {
			cx, cy := e.X+d[0], e.Y+d[1]
			if cx >= 0 && cx < mapWidth && cy >= 0 && cy < mapHeight && g.worldMap[cx][cy] != Wall {
				nx, ny = cx, cy
				break
			}
		}
	}

	if nx != e.X || ny != e.Y {
		e.Dir = dirFromDelta(nx-e.X, ny-e.Y)
		if !g.isEnemyAt(nx, ny) && !(nx == g.Player.X && ny == g.Player.Y) {
			e.X, e.Y = nx, ny
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
