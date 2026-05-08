package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// キャラクターの向きを示す小ドットをタイル上に描画する
func drawDirDot(screen *ebiten.Image, tileX, tileY int, d Dir, clr color.Color, camX, camY float64) {
	const dotSize = 6
	const half = (tileSize - 1) / 2
	ox, oy := 0, 0
	switch d {
	case DirUp:
		ox, oy = half-dotSize/2, 4
	case DirDown:
		ox, oy = half-dotSize/2, tileSize-1-dotSize-4
	case DirLeft:
		ox, oy = 4, half-dotSize/2
	case DirRight:
		ox, oy = tileSize-1-dotSize-4, half-dotSize/2
	}
	dot := ebiten.NewImage(dotSize, dotSize)
	dot.Fill(clr)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(tileX*tileSize+ox)-camX, float64(tileY*tileSize+oy)-camY)
	screen.DrawImage(dot, op)
}

func measureText(str string, face *text.GoTextFace) int {
	w, _ := text.Measure(str, face, 0)
	return int(w) + 1
}

// truncateText は str がmaxWidth px を超える場合に "…" を付けて切り詰める
func truncateText(str string, face *text.GoTextFace, maxWidth int) string {
	if measureText(str, face) <= maxWidth {
		return str
	}
	const ellipsis = "…"
	ellipsisW := measureText(ellipsis, face)
	runes := []rune(str)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		if measureText(string(runes), face)+ellipsisW <= maxWidth {
			return string(runes) + ellipsis
		}
	}
	return ellipsis
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

// drawPanel は角丸なしのパネルと枠線を描画するヘルパー
func drawPanel(screen *ebiten.Image, x, y, w, h int, bg, border color.RGBA) {
	panel := ebiten.NewImage(w, h)
	panel.Fill(bg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(panel, op)

	for _, r := range []struct{ x, y, w, h int }{
		{x, y, w, 2},
		{x, y + h - 2, w, 2},
		{x, y, 2, h},
		{x + w - 2, y, 2, h},
	} {
		b := ebiten.NewImage(r.w, r.h)
		b.Fill(border)
		bOp := &ebiten.DrawImageOptions{}
		bOp.GeoM.Translate(float64(r.x), float64(r.y))
		screen.DrawImage(b, bOp)
	}
}

// wrapText は str を face で描画したとき maxWidth px を超えないよう行分割する
func wrapText(str string, face *text.GoTextFace, maxWidth int) []string {
	var lines []string
	runes := []rune(str)
	start := 0
	for start < len(runes) {
		end := len(runes)
		lo, hi := start+1, len(runes)
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if measureText(string(runes[start:mid]), face) <= maxWidth {
				lo = mid
			} else {
				hi = mid - 1
			}
		}
		end = lo
		lines = append(lines, string(runes[start:end]))
		start = end
	}
	return lines
}
