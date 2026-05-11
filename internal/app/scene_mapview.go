package app

import "github.com/gentaman/myrogue/internal/core/world"

type MapViewScene struct {
	game *GameScene
}

func (s *MapViewScene) Update(input InputState) Scene {
	if input.Cancel || input.MapView {
		return s.game
	}
	return nil
}

func (s *MapViewScene) Draw(r Renderer) {
	r.Clear(0, 0, 0, 255)

	const cellSize = 4
	offsetX := (ScreenWidth - world.MapWidth*cellSize) / 2
	offsetY := (ScreenHeight - world.MapHeight*cellSize) / 2

	for x := 0; x < world.MapWidth; x++ {
		for y := 0; y < world.MapHeight; y++ {
			if !s.game.World.Explored[x][y] {
				continue
			}
			var clr Color
			switch s.game.World.Tiles[x][y] {
			case world.Wall:
				clr = Color{80, 80, 90, 255}
			case world.Floor:
				clr = Color{40, 40, 45, 255}
			case world.Stairs, world.StairsDown, world.StairsUp:
				clr = Color{0, 200, 200, 255}
			default:
				continue
			}
			if !s.game.World.Visible[x][y] {
				clr = Color{clr[0] / 2, clr[1] / 2, clr[2] / 2, 255}
			}
			r.DrawRect(offsetX+x*cellSize, offsetY+y*cellSize, cellSize, cellSize, clr)
		}
	}

	// Player position
	pPos := s.game.playerPos()
	px := offsetX + pPos.X*cellSize
	py := offsetY + pPos.Y*cellSize
	r.DrawRect(px, py, cellSize, cellSize, Color{255, 255, 100, 255})

	r.DrawText("マップ", 14, ScreenWidth/2, 12, Color{200, 200, 200, 255}, true)
	r.DrawText("M/Esc: 閉じる", 12, ScreenWidth/2, ScreenHeight-20, Color{100, 100, 100, 255}, true)
}
