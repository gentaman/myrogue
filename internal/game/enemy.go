package game

import (
	"fmt"
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

// Enemy は敵キャラクターを表す
type Enemy struct {
	x, y int
	dir  Dir
}

func (g *GameScene) isEnemyAt(x, y int) bool {
	for _, e := range g.enemies {
		if e.x == x && e.y == y {
			return true
		}
	}
	return false
}

// プレイヤーが敵を攻撃（移動先の敵を倒す）
func (g *GameScene) attackEnemy(x, y int) {
	for i, e := range g.enemies {
		if e.x == x && e.y == y {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
			g.message = "敵を倒した！"
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

// 敵がプレイヤーを視認できるか判定する。
// 同じ部屋にいるか、通路上で水平/垂直の直線上に壁なく並んでいる場合に true。
func (g *GameScene) canSeePlayer(e *Enemy) bool {
	er := g.roomOf(e.x, e.y)
	pr := g.roomOf(g.playerX, g.playerY)

	// 同じ部屋にいる
	if er != -1 && er == pr {
		return true
	}

	// 通路上での水平/垂直 LOS
	if e.x == g.playerX {
		minY, maxY := e.y, g.playerY
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY + 1; y < maxY; y++ {
			if g.worldMap[e.x][y] == Wall {
				return false
			}
		}
		return true
	}
	if e.y == g.playerY {
		minX, maxX := e.x, g.playerX
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX + 1; x < maxX; x++ {
			if g.worldMap[x][e.y] == Wall {
				return false
			}
		}
		return true
	}
	return false
}

// BFS で (startX,startY) から (goalX,goalY) への次の1マスを返す。
// 到達不能なら (startX,startY) を返す。
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
	// 経路を逆トレースして start の次のマスを返す
	cur := pos{goalX, goalY}
	for prev[cur.x][cur.y] != (pos{startX, startY}) {
		cur = prev[cur.x][cur.y]
		if cur.x == startX && cur.y == startY {
			return startX, startY
		}
	}
	return cur.x, cur.y
}

// 敵の移動（視界内ならBFS追跡、視界外はランダム移動。正面に隣接時は攻撃）
func (g *GameScene) moveEnemies() {
	dirList := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	for i := range g.enemies {
		e := &g.enemies[i]

		// 正面にプレイヤーがいるなら攻撃（移動より先に判定）
		fdx, fdy := e.dir.delta()
		if e.x+fdx == g.playerX && e.y+fdy == g.playerY {
			g.playerHP--
			playSFXHit()
			if g.playerHP <= 0 {
				g.playState = StateDead
				g.message = "力尽きた..."
				return
			}
			g.message = fmt.Sprintf("敵に攻撃された！ HP: %d", g.playerHP)
			continue
		}

		var nx, ny int
		if g.canSeePlayer(e) {
			// 視界あり: BFS 最短経路の次の1マスへ
			nx, ny = g.bfsNextStep(e.x, e.y, g.playerX, g.playerY)
		} else {
			// 視界なし: ランダム移動（通行可能なマスから選ぶ）
			candidates := dirList
			rand.Shuffle(len(candidates), func(a, b int) { candidates[a], candidates[b] = candidates[b], candidates[a] })
			nx, ny = e.x, e.y
			for _, d := range candidates {
				cx, cy := e.x+d[0], e.y+d[1]
				if cx >= 0 && cx < mapWidth && cy >= 0 && cy < mapHeight && g.worldMap[cx][cy] != Wall {
					nx, ny = cx, cy
					break
				}
			}
		}

		if nx == e.x && ny == e.y {
			continue
		}

		// 向きを移動方向に更新
		e.dir = dirFromDelta(nx-e.x, ny-e.y)

		// 移動先にプレイヤーや他の敵がいなければ移動
		if !g.isEnemyAt(nx, ny) && !(nx == g.playerX && ny == g.playerY) {
			e.x = nx
			e.y = ny
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
