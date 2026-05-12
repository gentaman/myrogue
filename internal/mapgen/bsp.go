package mapgen

import (
	"math/rand"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/world"
)

const minLeafSize = 8

type bspLeaf struct {
	x, y, w, h  int
	left, right *bspLeaf
	room        *world.Room
}

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
		l.left = &bspLeaf{x: l.x, y: l.y, w: l.w, h: split}
		l.right = &bspLeaf{x: l.x, y: l.y + split, w: l.w, h: l.h - split}
	} else {
		l.left = &bspLeaf{x: l.x, y: l.y, w: split, h: l.h}
		l.right = &bspLeaf{x: l.x + split, y: l.y, w: l.w - split, h: l.h}
	}
	return true
}

func (l *bspLeaf) createRooms(gm *world.GameMap, rng *rand.Rand) {
	if l.left != nil || l.right != nil {
		if l.left != nil {
			l.left.createRooms(gm, rng)
		}
		if l.right != nil {
			l.right.createRooms(gm, rng)
		}
		return
	}
	w := rng.Intn(l.w-5) + 4
	h := rng.Intn(l.h-5) + 4
	x := l.x + rng.Intn(l.w-w-1) + 1
	y := l.y + rng.Intn(l.h-h-1) + 1
	l.room = &world.Room{X: x, Y: y, W: w, H: h}
	for rx := x; rx < x+w; rx++ {
		for ry := y; ry < y+h; ry++ {
			gm.Tiles[rx][ry] = world.Floor
		}
	}
	gm.Rooms = append(gm.Rooms, *l.room)
}

func (l *bspLeaf) createCorridors(gm *world.GameMap, rng *rand.Rand) {
	if l.left != nil && l.right != nil {
		l.left.createCorridors(gm, rng)
		l.right.createCorridors(gm, rng)
		r1, r2 := l.left.getRoom(), l.right.getRoom()
		if r1 != nil && r2 != nil {
			digCorridor(gm, rng, r1.CenterX(), r1.CenterY(), r2.CenterX(), r2.CenterY())
		}
	}
}

func (l *bspLeaf) getRoom() *world.Room {
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

func digCorridor(gm *world.GameMap, rng *rand.Rand, x1, y1, x2, y2 int) {
	if rng.Intn(2) == 0 {
		digHorizontal(gm, x1, x2, y1)
		digVertical(gm, y1, y2, x2)
	} else {
		digVertical(gm, y1, y2, x1)
		digHorizontal(gm, x1, x2, y2)
	}
}

func digHorizontal(gm *world.GameMap, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if gm.Tiles[x][y] == world.Wall {
			gm.Tiles[x][y] = world.Floor
		}
	}
}

func digVertical(gm *world.GameMap, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if gm.Tiles[x][y] == world.Wall {
			gm.Tiles[x][y] = world.Floor
		}
	}
}

func nearSpecialTile(gm *world.GameMap, x, y int) bool {
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			nx, ny := x+dx, y+dy
			if gm.InBounds(nx, ny) {
				t := gm.Tiles[nx][ny]
				if t == world.Stairs || t == world.StairsUp || t == world.StairsDown {
					return true
				}
			}
		}
	}
	return false
}

func Generate(floor int, floorDef *content.FloorDef, registry *content.Registry, seed int64, checker content.DropConditionChecker) *world.GameMap {
	rng := rand.New(rand.NewSource(seed))
	gm := &world.GameMap{
		Seed:  seed,
		Floor: floor,
	}
	for x := 0; x < world.MapWidth; x++ {
		for y := 0; y < world.MapHeight; y++ {
			gm.Tiles[x][y] = world.Wall
		}
	}

	root := &bspLeaf{x: 0, y: 0, w: world.MapWidth, h: world.MapHeight}
	leaves := []*bspLeaf{root}
	didSplit := true
	for didSplit && len(leaves) < floorDef.MaxRooms {
		didSplit = false
		var newLeaves []*bspLeaf
		for _, l := range leaves {
			if l.left == nil && l.right == nil {
				if len(leaves)+len(newLeaves) < floorDef.MinRooms || rng.Float64() > 0.3 {
					if l.split(rng) {
						newLeaves = append(newLeaves, l.left, l.right)
						didSplit = true
					}
				}
			}
		}
		leaves = append(leaves, newLeaves...)
	}
	root.createRooms(gm, rng)
	root.createCorridors(gm, rng)

	if floor == 0 && len(gm.Rooms) > 0 {
		r := gm.Rooms[0]
		gm.Tiles[r.CenterX()][r.CenterY()] = world.Stairs
	} else if len(gm.Rooms) > 0 {
		r := gm.Rooms[0]
		gm.Tiles[r.CenterX()][r.CenterY()] = world.StairsUp
	}

	if floor < registry.FloorCount()-1 && len(gm.Rooms) >= 2 {
		for attempt := 0; attempt < 200; attempt++ {
			roomIdx := 1 + rng.Intn(len(gm.Rooms)-1)
			r := gm.Rooms[roomIdx]
			dx, dy := r.X+rng.Intn(r.W), r.Y+rng.Intn(r.H)
			if gm.Tiles[dx][dy] == world.Floor {
				gm.Tiles[dx][dy] = world.StairsDown
				break
			}
		}
	}

	placeItems(gm, floorDef, registry, rng, checker)
	return gm
}

func placeItems(gm *world.GameMap, floorDef *content.FloorDef, registry *content.Registry, rng *rand.Rand, checker content.DropConditionChecker) {
	// 配置予定アイテムのリスト
	var itemsToPlace []string

	// 条件付き（または無条件）アイテムの評価
	for _, ci := range floorDef.ConditionalItems {
		if ci.Condition == "" || ci.Condition == "always" {
			itemsToPlace = append(itemsToPlace, ci.ID)
		} else if checker != nil && checker.CheckCondition(ci.Condition, ci.ID) {
			itemsToPlace = append(itemsToPlace, ci.ID)
		}
	}

	// 最下層（最後のフロア）なら「treasure」を必ず追加（既に持っていない場合のみ、かつ定義に含まれていない場合）
	if gm.Floor == registry.FloorCount()-1 {
		hasTreasure := false
		for _, id := range itemsToPlace {
			if id == "treasure" {
				hasTreasure = true
				break
			}
		}
		if !hasTreasure {
			// 所持チェック
			if checker == nil || checker.CheckCondition("unique", "treasure") {
				itemsToPlace = append(itemsToPlace, "treasure")
			}
		}
	}

	// 特殊アイテム（条件クリア分）の配置
	for _, itemID := range itemsToPlace {
		for attempt := 0; attempt < 100; attempt++ {
			roomIdx := rng.Intn(len(gm.Rooms))
			r := gm.Rooms[roomIdx]
			ix, iy := r.X+rng.Intn(r.W), r.Y+rng.Intn(r.H)
			if gm.Tiles[ix][iy] != world.Floor || nearSpecialTile(gm, ix, iy) {
				continue
			}
			def, ok := registry.GetItemDef(itemID)
			if !ok {
				break
			}
			entry := ItemEntryFromDef(def, gm.Seed, gm.Floor)
			gm.Items = append(gm.Items, world.MapItem{
				X: ix, Y: iy,
				Inventory: []component.ItemEntry{entry},
			})
			break
		}
	}

	// ランダムアイテムの配置
	weights := make([]int, len(floorDef.ItemPool))
	for i, entry := range floorDef.ItemPool {
		weights[i] = entry.Weight
	}
	for i := 0; i < floorDef.ItemCount; i++ {
		for attempt := 0; attempt < 100; attempt++ {
			roomIdx := rng.Intn(len(gm.Rooms))
			r := gm.Rooms[roomIdx]
			ix, iy := r.X+rng.Intn(r.W), r.Y+rng.Intn(r.H)
			if gm.Tiles[ix][iy] != world.Floor || nearSpecialTile(gm, ix, iy) {
				continue
			}
			poolIdx := content.WeightedChoice(rng, weights)
			if poolIdx >= 0 {
				itemID := floorDef.ItemPool[poolIdx].ID
				def, ok := registry.GetItemDef(itemID)
				if !ok {
					break
				}
				entry := ItemEntryFromDef(def, gm.Seed, gm.Floor)
				gm.Items = append(gm.Items, world.MapItem{
					X: ix, Y: iy,
					Inventory: []component.ItemEntry{entry},
				})
			}
			break
		}
	}
}

func ItemEntryFromDef(def *content.ItemDef, seed int64, floor int) component.ItemEntry {
	s, f := int64(0), 0
	if def.FloorBound {
		s, f = seed, floor
	}
	return component.ItemEntry{
		DefID:         def.ID,
		Count:         1,
		Durability:    def.Durability,
		ObtainedSeed:  s,
		ObtainedFloor: f,
	}
}
