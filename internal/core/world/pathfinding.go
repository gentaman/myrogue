package world

type BlockedFunc func(x, y int) bool

func BFSNextStep(gm *GameMap, startX, startY, goalX, goalY int, blocked BlockedFunc) (int, int) {
	if startX == goalX && startY == goalY {
		return startX, startY
	}
	type pos struct{ x, y int }
	var prev [MapWidth][MapHeight]pos
	var visited [MapWidth][MapHeight]bool
	visited[startX][startY] = true
	queue := []pos{{startX, startY}}
	dirs := [4]pos{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	found := false
	for len(queue) > 0 && !found {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range dirs {
			nx, ny := cur.x+d.x, cur.y+d.y
			if !gm.InBounds(nx, ny) || visited[nx][ny] || gm.Tiles[nx][ny] == Wall {
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

func BFSFleeStep(gm *GameMap, startX, startY, fearX, fearY int, blocked BlockedFunc) (int, int) {
	type pos struct{ x, y int }
	dirs := [4]pos{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	bestX, bestY := startX, startY
	maxDist := (startX-fearX)*(startX-fearX) + (startY-fearY)*(startY-fearY)
	for _, d := range dirs {
		nx, ny := startX+d.x, startY+d.y
		if !gm.InBounds(nx, ny) || gm.Tiles[nx][ny] == Wall {
			continue
		}
		if blocked != nil && blocked(nx, ny) {
			continue
		}
		dist := (nx-fearX)*(nx-fearX) + (ny-fearY)*(ny-fearY)
		if dist > maxDist {
			maxDist = dist
			bestX, bestY = nx, ny
		}
	}
	return bestX, bestY
}

func CanSee(gm *GameMap, fromX, fromY, toX, toY int) bool {
	dx := Abs(toX - fromX)
	dy := Abs(toY - fromY)
	sx := -1
	if fromX < toX {
		sx = 1
	}
	sy := -1
	if fromY < toY {
		sy = 1
	}
	err := dx - dy

	x, y := fromX, fromY
	for {
		if x == toX && y == toY {
			return true
		}
		if x != fromX || y != fromY {
			if gm.Tiles[x][y] == Wall {
				return false
			}
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
