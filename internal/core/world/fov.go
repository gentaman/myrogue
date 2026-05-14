package world

func UpdateVisibility(gm *GameMap, px, py int) {
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			gm.Visible[x][y] = false
		}
	}
	roomIdx := gm.RoomOf(px, py)
	if roomIdx >= 0 {
		r := gm.Rooms[roomIdx]
		for x := r.X - 1; x < r.X+r.W+1; x++ {
			for y := r.Y - 1; y < r.Y+r.H+1; y++ {
				if gm.InBounds(x, y) {
					gm.Visible[x][y] = true
				}
			}
		}
	} else {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				x, y := px+dx, py+dy
				if gm.InBounds(x, y) {
					gm.Visible[x][y] = true
				}
			}
		}
	}
	for x := 0; x < MapWidth; x++ {
		for y := 0; y < MapHeight; y++ {
			if gm.Visible[x][y] {
				gm.Explored[x][y] = true
			}
		}
	}
}
