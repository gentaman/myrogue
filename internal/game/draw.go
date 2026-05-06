package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// キャラクターの向きを示す小ドットをタイル上に描画する
func drawDirDot(screen *ebiten.Image, tileX, tileY int, d Dir, clr color.Color) {
	const dotSize = 3
	const half = (tileSize - 1) / 2
	ox, oy := 0, 0
	switch d {
	case DirUp:
		ox, oy = half-dotSize/2, 1
	case DirDown:
		ox, oy = half-dotSize/2, tileSize-1-dotSize-1
	case DirLeft:
		ox, oy = 1, half-dotSize/2
	case DirRight:
		ox, oy = tileSize-1-dotSize-1, half-dotSize/2
	}
	dot := ebiten.NewImage(dotSize, dotSize)
	dot.Fill(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(tileX*tileSize+ox), float64(tileY*tileSize+oy))
	screen.DrawImage(dot, op)
}

func drawText(screen *ebiten.Image, str string, face *text.GoTextFace, x, y int, clr color.Color, center bool) {
	op := &text.DrawOptions{}
	if center {
		op.PrimaryAlign = text.AlignCenter
	}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, face, op)
}
