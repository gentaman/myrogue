package game

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"time"
)

// TileType はマップのタイル種別を表す
type TileType int

const (
	Wall TileType = iota
	Floor
	Stairs     // 上り階段（フロア0の出口）
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
	x, y, w, h  int
	left, right *bspLeaf
	room        *Room
}

const minLeafSize = 8

func (l *bspLeaf) split(rng *rand.Rand) bool {
	if l.left != nil || l.right != nil {
		return false
	}
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
		r1, r2 := l.left.getRoom(), l.right.getRoom()
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

type poolEntry struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
}

type floorDef struct {
	Floor      int         `json:"floor"`
	MinRooms   int         `json:"min_rooms"`
	MaxRooms   int         `json:"max_rooms"`
	EnemyCount int         `json:"enemy_count"`
	ItemCount  int         `json:"item_count"`
	EnemyPool  []poolEntry `json:"enemy_pool"`
	ItemPool   []poolEntry `json:"item_pool"`
}

var (
	//go:embed assets/floors.json
	floorsJSON []byte
	floorDefs  []floorDef
)

func init() {
	if err := json.Unmarshal(floorsJSON, &floorDefs); err != nil {
		panic(err)
	}
}

func (g *GameScene) generateMap() {
	seed := time.Now().UnixNano()
	g.mapSeed = seed
	rng := rand.New(rand.NewSource(seed))
	fDef := &floorDefs[g.floor]
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.worldMap[x][y] = Wall
		}
	}
	g.rooms, g.mapItems = nil, nil
	root := &bspLeaf{0, 0, mapWidth, mapHeight, nil, nil, nil}
	leaves := []*bspLeaf{root}
	didSplit := true
	for didSplit && len(leaves) < fDef.MaxRooms {
		didSplit = false
		newLeaves := []*bspLeaf{}
		for _, l := range leaves {
			if l.left == nil && l.right == nil {
				if len(leaves)+len(newLeaves) < fDef.MinRooms || rng.Float64() > 0.3 {
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
		g.Player.X = g.rooms[0].centerX()
		g.Player.Y = g.rooms[0].centerY()
		if g.floor == 0 {
			g.worldMap[g.Player.X][g.Player.Y] = Stairs
		} else {
			g.worldMap[g.Player.X][g.Player.Y] = StairsUp
		}
	}
	if g.floor == len(floorDefs)-1 && !g.HasTreasure() {
		for attempt := 0; attempt < 100; attempt++ {
			roomIdx := rng.Intn(len(g.rooms))
			r := g.rooms[roomIdx]
			tx, ty := r.x+rng.Intn(r.w), r.y+rng.Intn(r.h)
			if g.worldMap[tx][ty] == Floor && !g.nearSpecialTile(tx, ty) {
				g.mapItems = append(g.mapItems, MapItem{
					X: tx, Y: ty,
					Inventory: []InventoryEntry{{
						kind:          itemIDMap["treasure"],
						count:         1,
						Durability:    itemDefs[itemIDMap["treasure"]].durability,
						obtainedSeed:  g.mapSeed,
						obtainedFloor: g.floor,
					}},
				})
				break
			}
		}
	}
	if g.floor < len(floorDefs)-1 && len(g.rooms) >= 2 {
		for attempt := 0; attempt < 200; attempt++ {
			roomIdx := 1 + rng.Intn(len(g.rooms)-1)
			r := g.rooms[roomIdx]
			dx, dy := r.x+rng.Intn(r.w), r.y+rng.Intn(r.h)
			if g.worldMap[dx][dy] == Floor {
				g.worldMap[dx][dy] = StairsDown
				break
			}
		}
	}
	g.maxEnemies = fDef.EnemyCount
	itemWeights := make([]int, len(fDef.ItemPool))
	for i, entry := range fDef.ItemPool {
		itemWeights[i] = entry.Weight
	}
	for i := 0; i < fDef.ItemCount; i++ {
		for attempt := 0; attempt < 100; attempt++ {
			roomIdx := rng.Intn(len(g.rooms))
			r := g.rooms[roomIdx]
			ix, iy := r.x+rng.Intn(r.w), r.y+rng.Intn(r.h)
			if g.worldMap[ix][iy] != Floor || g.nearSpecialTile(ix, iy) {
				continue
			}
			poolIdx := weightedChoice(rng, fDef.ItemPool, itemWeights)
			if poolIdx >= 0 {
				itemID := fDef.ItemPool[poolIdx].ID
				kind := itemIDMap[itemID]
				g.mapItems = append(g.mapItems, MapItem{
					X: ix, Y: iy,
					Inventory: []InventoryEntry{{
						kind:          kind,
						count:         1,
						Durability:    itemDefs[kind].durability,
						obtainedSeed:  g.mapSeed,
						obtainedFloor: g.floor,
					}},
				})
			}
			break
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

func (g *GameScene) roomOf(x, y int) int {
	for i, r := range g.rooms {
		if x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h {
			return i
		}
	}
	return -1
}

func (g *GameScene) playerRoom() int {
	return g.roomOf(g.Player.X, g.Player.Y)
}

func (g *GameScene) nearSpecialTile(x, y int) bool {
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < mapWidth && ny >= 0 && ny < mapHeight {
				t := g.worldMap[nx][ny]
				if t == Stairs || t == StairsUp || t == StairsDown {
					return true
				}
			}
		}
	}
	return false
}

func (g *GameScene) spawnInitialEnemies() {
	playerRoomIdx := g.playerRoom()
	count := len(g.rooms) / 2
	for i := 0; i < count; i++ {
		g.trySpawnEnemy(playerRoomIdx)
	}
}

func (g *GameScene) trySpawnEnemyPerTurn() {
	if len(g.enemies) >= g.maxEnemies {
		return
	}
	if rand.Intn(100) >= 30 {
		return
	}
	g.trySpawnEnemy(g.playerRoom())
}

func (g *GameScene) trySpawnEnemy(playerRoomIdx int) {
	fDef := &floorDefs[g.floor]
	if len(fDef.EnemyPool) == 0 {
		return
	}
	weights := make([]int, len(fDef.EnemyPool))
	for i, entry := range fDef.EnemyPool {
		weights[i] = entry.Weight
	}
	for attempt := 0; attempt < 50; attempt++ {
		roomIdx := rand.Intn(len(g.rooms))
		if roomIdx == playerRoomIdx {
			continue
		}
		r := g.rooms[roomIdx]
		ex, ey := r.x+rand.Intn(r.w), r.y+rand.Intn(r.h)
		if g.worldMap[ex][ey] != Floor || g.nearSpecialTile(ex, ey) || g.isEnemyAt(ex, ey) {
			continue
		}
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(attempt)))
		poolIdx := weightedChoice(rng, fDef.EnemyPool, weights)
		kindIdx := enemyIDMap[fDef.EnemyPool[poolIdx].ID]
		def := &enemyDefs[kindIdx]

		enemyInv := []InventoryEntry{}
		// 確率で木の武器や盾を持たせる
		if rand.Intn(2) == 0 {
			weaponKind := itemIDMap["weapon_wood"]
			if g.floor > 0 && rand.Intn(2) == 0 {
				weaponKind = itemIDMap["weapon_iron"]
			}
			enemyInv = append(enemyInv, InventoryEntry{
				kind:       weaponKind,
				count:      1,
				Durability: itemDefs[weaponKind].durability,
				Equipped:   true,
			})
		}

		enemyID := time.Now().UnixNano() + int64(len(g.enemies))
		g.enemies = append(g.enemies, Enemy{
			Actor: Actor{
				ID:        enemyID,
				X:         ex,
				Y:         ey,
				HP:        def.HP,
				MaxHP:     def.HP,
				Level:     1 + g.floor*2, // フロアに応じた初期レベル
				XP:        0,
				XPToNext:  10, // 固定または計算
				Str:       def.Str,
				Wis:       def.Wis,
				Fai:       def.Fai,
				Vit:       def.Vit,
				Agi:       def.Agi,
				Luk:       def.Luk,
				Dir:       DirDown,
				Race:      def.Race,
				Inventory: enemyInv,
				Relations: make(map[int64]int),
			},
			kind:     EnemyKind(kindIdx),
			rewardXP: def.XP,
		})
		return
	}
}
