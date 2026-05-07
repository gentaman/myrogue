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
	Stairs // 上り階段（フロア0の出口）
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

type bspLeaf struct {
	x, y, w, h int
	left, right *bspLeaf
	room       *Room
}

const minLeafSize = 8

func (l *bspLeaf) split(rng *rand.Rand) bool {
	if l.left != nil || l.right != nil {
		return false
	}

	// 分割方向の決定
	splitH := rng.Intn(2) == 0
	if l.w > l.h && float64(l.w)/float64(l.h) >= 1.25 {
		splitH = false
	} else if l.h > l.w && float64(l.h)/float64(l.w) >= 1.25 {
		splitH = true
	}

	max := l.w
	if splitH {
		max = l.h
	}
	if max <= minLeafSize*2 {
		return false
	}

	splitRange := max - minLeafSize*2
	var split int
	if splitRange > 0 {
		split = rng.Intn(splitRange) + minLeafSize
	} else {
		split = minLeafSize
	}

	if splitH {
		l.left = &bspLeaf{l.x, l.y, l.w, split, nil, nil, nil}
		l.right = &bspLeaf{l.x, l.y + split, l.w, l.h - split, nil, nil, nil}
	} else {
		l.left = &bspLeaf{l.x, l.y, split, l.h, nil, nil, nil}
		l.right = &bspLeaf{l.x + split, l.y, l.w - split, l.h, nil, nil, nil}
	}
	return true
}

func (l *bspLeaf) createRooms(g *GameScene, rng *rand.Rand) {
	if l.left != nil || l.right != nil {
		if l.left != nil {
			l.left.createRooms(g, rng)
		}
		if l.right != nil {
			l.right.createRooms(g, rng)
		}
	} else {
		// 部屋のサイズ: 最小3x3, リーフサイズ-パディングまで
		w := rng.Intn(l.w-5) + 4
		h := rng.Intn(l.h-5) + 4
		x := l.x + rng.Intn(l.w-w-1) + 1
		y := l.y + rng.Intn(l.h-h-1) + 1
		l.room = &Room{x, y, w, h}

		for rx := x; rx < x+w; rx++ {
			for ry := y; ry < y+h; ry++ {
				g.worldMap[rx][ry] = Floor
			}
		}
		g.rooms = append(g.rooms, *l.room)
	}
}

func (l *bspLeaf) createCorridors(g *GameScene, rng *rand.Rand) {
	if l.left != nil && l.right != nil {
		l.left.createCorridors(g, rng)
		l.right.createCorridors(g, rng)

		r1 := l.left.getRoom()
		r2 := l.right.getRoom()
		if r1 != nil && r2 != nil {
			g.digCorridor(rng, r1.centerX(), r1.centerY(), r2.centerX(), r2.centerY())
		}
	}
}

func (l *bspLeaf) getRoom() *Room {
	if l.room != nil {
		return l.room
	}
	if l.left != nil {
		if r := l.left.getRoom(); r != nil {
			return r
		}
	}
	if l.right != nil {
		if r := l.right.getRoom(); r != nil {
			return r
		}
	}
	return nil
}

func (g *GameScene) generateMap() {
	seed := time.Now().UnixNano()
	g.mapSeed = seed
	rng := rand.New(rand.NewSource(seed))

	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.worldMap[x][y] = Wall
		}
	}

	g.rooms = nil
	g.mapItems = nil

	root := &bspLeaf{0, 0, mapWidth, mapHeight, nil, nil, nil}
	leaves := []*bspLeaf{root}

	didSplit := true
	for didSplit {
		didSplit = false
		newLeaves := []*bspLeaf{}
		for _, l := range leaves {
			if l.left == nil && l.right == nil {
				if l.w > minLeafSize*2 || l.h > minLeafSize*2 || rng.Float64() > 0.25 {
					if l.split(rng) {
						newLeaves = append(newLeaves, l.left, l.right)
						didSplit = true
					}
				}
			}
		}
		leaves = append(leaves, newLeaves...)
	}

	root.createRooms(g, rng)
	root.createCorridors(g, rng)

	if len(g.rooms) > 0 {
		g.playerX = g.rooms[0].centerX()
		g.playerY = g.rooms[0].centerY()
		if g.floor == 0 {
			g.worldMap[g.playerX][g.playerY] = Stairs
		} else {
			g.worldMap[g.playerX][g.playerY] = StairsUp
		}
	}

	// 最下層のみ宝を配置
	if g.floor == maxFloor-1 {
		for {
			tx := rng.Intn(mapWidth)
			ty := rng.Intn(mapHeight)
			if g.worldMap[tx][ty] == Floor {
				g.worldMap[tx][ty] = Treasure
				break
			}
		}
	}

	// 最下層以外は下り階段を配置（上り階段のある rooms[0] 以外の部屋に限定）
	if g.floor < maxFloor-1 && len(g.rooms) >= 2 {
		for attempt := 0; attempt < 200; attempt++ {
			roomIdx := 1 + rng.Intn(len(g.rooms)-1)
			r := g.rooms[roomIdx]
			dx := r.x + rng.Intn(r.w)
			dy := r.y + rng.Intn(r.h)
			if g.worldMap[dx][dy] == Floor {
				g.worldMap[dx][dy] = StairsDown
				break
			}
		}
	}

	g.maxEnemies = len(g.rooms)

	weights := []int{}

	for i := 0; i < int(itemKindCount)-1; i++ {
		weights = append(weights, -itemDefs[i].rarity)
	}

	// 各部屋に確率でアイテムを配置
	for _, r := range g.rooms {
		if rng.Intn(2) == 0 {
			for attempt := 0; attempt < 30; attempt++ {
				ix := r.x + rng.Intn(r.w)
				iy := r.y + rng.Intn(r.h)
				if g.worldMap[ix][iy] != Floor {
					continue
				}
				idx := weightedChoice(rng, itemDefs, weights)
				if idx >= 0 {
					kind := ItemKind(idx)
					g.mapItems = append(g.mapItems, MapItem{x: ix, y: iy, kind: kind, obtainedSeed: g.mapSeed, obtainedFloor: g.floor})
				}
				break
			}
		}
	}
}

func (g *GameScene) digCorridor(rng *rand.Rand, x1, y1, x2, y2 int) {
	if rng.Intn(2) == 0 {
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
			if y > 0 {
				g.worldMap[x][y-1] = Wall
			}
			if y < mapHeight-1 {
				g.worldMap[x][y+1] = Wall
			}
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
			if x > 0 {
				g.worldMap[x-1][y] = Wall
			}
			if x < mapWidth-1 {
				g.worldMap[x+1][y] = Wall
			}
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
