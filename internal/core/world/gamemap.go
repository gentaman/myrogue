package world

import "github.com/gentaman/myrogue/internal/core/component"

const (
	MapWidth  = 40
	MapHeight = 25
)

type Room struct {
	X, Y, W, H int
}

func (r Room) CenterX() int { return r.X + r.W/2 }
func (r Room) CenterY() int { return r.Y + r.H/2 }

type MapItem struct {
	X, Y      int
	Inventory []component.ItemEntry
}

type GameMap struct {
	Tiles    [MapWidth][MapHeight]TileType
	Explored [MapWidth][MapHeight]bool
	Visible  [MapWidth][MapHeight]bool
	Rooms    []Room
	Seed     int64
	Floor    int
	Items    []MapItem
}

func (m *GameMap) InBounds(x, y int) bool {
	return x >= 0 && x < MapWidth && y >= 0 && y < MapHeight
}

func (m *GameMap) IsPassable(x, y int) bool {
	if !m.InBounds(x, y) {
		return false
	}
	return m.Tiles[x][y] != Wall
}

func (m *GameMap) RoomOf(x, y int) int {
	for i, r := range m.Rooms {
		if x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
			return i
		}
	}
	return -1
}
