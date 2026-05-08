package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type MapScene struct {
	game *GameScene
}

func (s *MapScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyM) {
		return s.game, nil
	}
	return nil, nil
}

func (s *MapScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	const scale = 0.25
	const drawW = float64(mapWidth) * tileSize * scale
	const drawH = float64(mapHeight) * tileSize * scale
	offsetX := (screenWidth - drawW) / 2
	offsetY := (screenHeight - drawH) / 2

	// マップ描画
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			if !s.game.explored[x][y] {
				continue
			}

			var clr color.Color
			switch s.game.worldMap[x][y] {
			case Wall:
				clr = color.RGBA{100, 100, 110, 255}
			case Floor:
				clr = color.RGBA{60, 60, 65, 255}
			case Stairs:
				clr = color.RGBA{0, 255, 255, 255}
			case StairsDown:
				clr = color.RGBA{255, 140, 0, 255}
			case StairsUp:
				clr = color.RGBA{0, 220, 180, 255}
			default:
				clr = color.RGBA{40, 40, 40, 255}
			}

			rect := ebiten.NewImage(tileSize, tileSize)
			rect.Fill(clr)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(offsetX+float64(x)*tileSize*scale, offsetY+float64(y)*tileSize*scale)
			screen.DrawImage(rect, op)
		}
	}

	// プレイヤー位置
	pRect := ebiten.NewImage(tileSize, tileSize)
	pRect.Fill(color.RGBA{255, 100, 100, 255})
	pOp := &ebiten.DrawImageOptions{}
	pOp.GeoM.Scale(scale, scale)
	pOp.GeoM.Translate(offsetX+float64(s.game.Player.X)*tileSize*scale, offsetY+float64(s.game.Player.Y)*tileSize*scale)
	screen.DrawImage(pRect, pOp)

	drawText(screen, "全体マップ (M/Esc: 戻る)", fontFace12, screenWidth/2, screenHeight-20, color.RGBA{255, 255, 255, 255}, true)
}
