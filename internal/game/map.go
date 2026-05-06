package game

import (
	"math/rand"
	"time"
)

// TileType はマップのタイル種別を表す
type TileType int

const (
	Wall TileType = iota
	Floor
	Stairs     // 上り階段（フロア0の出口）
	Treasure
	StairsDown // 下り階段（次のフロアへ）
	StairsUp   // 上り階段（前のフロアへ、フロア1・2に配置）
)

// Room はダンジョンの部屋を表す
type Room struct {
	x, y, w, h int
}

func (r Room) centerX() int { return r.x + r.w/2 }
func (r Room) centerY() int { return r.y + r.h/2 }

func (g *GameScene) generateMap() {
	rand.Seed(time.Now().UnixNano())

	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.worldMap[x][y] = Wall
		}
	}

	for attempt := 0; len(g.rooms) < 5 && attempt < 200; attempt++ {
		w := rand.Intn(6) + 4
		h := rand.Intn(6) + 4
		x := rand.Intn(mapWidth-w-2) + 1
		y := rand.Intn(mapHeight-h-2) + 1

		// 既存の部屋と1マスのマージンを含めて重複しないか確認
		overlaps := false
		for _, r := range g.rooms {
			if x-1 < r.x+r.w && x+w+1 > r.x && y-1 < r.y+r.h && y+h+1 > r.y {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}

		for rx := x; rx < x+w; rx++ {
			for ry := y; ry < y+h; ry++ {
				g.worldMap[rx][ry] = Floor
			}
		}
		g.rooms = append(g.rooms, Room{x, y, w, h})
	}

	for i := 1; i < len(g.rooms); i++ {
		g.digCorridor(g.rooms[i-1].centerX(), g.rooms[i-1].centerY(), g.rooms[i].centerX(), g.rooms[i].centerY())
	}

	if len(g.rooms) > 0 {
		g.playerX = g.rooms[0].centerX()
		g.playerY = g.rooms[0].centerY()
		// フロア0のみ上り出口（Stairs）を配置、それ以外は上り階段（StairsUp）
		if g.floor == 0 {
			g.worldMap[g.playerX][g.playerY] = Stairs
		} else {
			g.worldMap[g.playerX][g.playerY] = StairsUp
		}
	}

	// 最下層のみ宝を配置
	if g.floor == maxFloor-1 {
		for {
			tx := rand.Intn(mapWidth)
			ty := rand.Intn(mapHeight)
			if g.worldMap[tx][ty] == Floor {
				g.worldMap[tx][ty] = Treasure
				break
			}
		}
	}

	// 最下層以外は下り階段を配置（上り階段のある rooms[0] 以外の部屋に限定）
	if g.floor < maxFloor-1 && len(g.rooms) >= 2 {
		for attempt := 0; attempt < 200; attempt++ {
			// rooms[1] 以降からランダムに選ぶ
			roomIdx := 1 + rand.Intn(len(g.rooms)-1)
			r := g.rooms[roomIdx]
			dx := r.x + rand.Intn(r.w)
			dy := r.y + rand.Intn(r.h)
			if g.worldMap[dx][dy] == Floor {
				g.worldMap[dx][dy] = StairsDown
				break
			}
		}
	}

	g.maxEnemies = len(g.rooms)
}

func (g *GameScene) digCorridor(x1, y1, x2, y2 int) {
	if rand.Intn(2) == 0 {
		g.digHorizontal(x1, x2, y1)
		g.digVertical(y1, y2, x2)
	} else {
		g.digVertical(y1, y2, x1)
		g.digHorizontal(x1, x2, y2)
	}
}

func (g *GameScene) digHorizontal(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if g.worldMap[x][y] == Wall {
			g.worldMap[x][y] = Floor
		}
	}
}

func (g *GameScene) digVertical(y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if g.worldMap[x][y] == Wall {
			g.worldMap[x][y] = Floor
		}
	}
}

// 指定座標がいる部屋のインデックスを返す（通路上なら-1）
func (g *GameScene) roomOf(x, y int) int {
	for i, r := range g.rooms {
		if x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h {
			return i
		}
	}
	return -1
}

// プレイヤーがいる部屋のインデックスを返す（通路上なら-1）
func (g *GameScene) playerRoom() int {
	return g.roomOf(g.playerX, g.playerY)
}

// 指定座標が階段または宝の周囲3マス以内か
func (g *GameScene) nearSpecialTile(x, y int) bool {
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < mapWidth && ny >= 0 && ny < mapHeight {
				t := g.worldMap[nx][ny]
				if t == Stairs || t == StairsUp || t == StairsDown || t == Treasure {
					return true
				}
			}
		}
	}
	return false
}

// 初期敵配置
func (g *GameScene) spawnInitialEnemies() {
	playerRoomIdx := g.playerRoom()
	count := len(g.rooms) / 2
	for i := 0; i < count; i++ {
		g.trySpawnEnemy(playerRoomIdx)
	}
}

// 毎ターン敵の追加生成を試みる
func (g *GameScene) trySpawnEnemyPerTurn() {
	if len(g.enemies) >= g.maxEnemies {
		return
	}
	// 一定確率で生成（毎ターン30%）
	if rand.Intn(100) >= 30 {
		return
	}
	g.trySpawnEnemy(g.playerRoom())
}

// プレイヤーと同じ部屋以外の床にランダムに1体生成
func (g *GameScene) trySpawnEnemy(playerRoomIdx int) {
	for attempt := 0; attempt < 50; attempt++ {
		roomIdx := rand.Intn(len(g.rooms))
		if roomIdx == playerRoomIdx {
			continue
		}
		r := g.rooms[roomIdx]
		ex := r.x + rand.Intn(r.w)
		ey := r.y + rand.Intn(r.h)
		if g.worldMap[ex][ey] != Floor {
			continue
		}
		if g.nearSpecialTile(ex, ey) {
			continue
		}
		if g.isEnemyAt(ex, ey) {
			continue
		}
		g.enemies = append(g.enemies, Enemy{x: ex, y: ey})
		return
	}
}
