package ebiten

import (
	"image"
	"image/color"
	"strconv"

	ebt "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/gentaman/myrogue/internal/app"
)

type EbitenRenderer struct {
	screen *ebt.Image
}

func (r *EbitenRenderer) ScreenSize() (int, int) {
	return screenWidth, screenHeight
}

func (r *EbitenRenderer) Clear(red, g, b, a uint8) {
	r.screen.Fill(color.RGBA{red, g, b, a})
}

func (r *EbitenRenderer) DrawRect(x, y, w, h int, clr app.Color) {
	rect := ebt.NewImage(w, h)
	rect.Fill(color.RGBA{clr[0], clr[1], clr[2], clr[3]})
	op := &ebt.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	r.screen.DrawImage(rect, op)
}

func (r *EbitenRenderer) DrawSprite(name string, frame, dir int, x, y float64) {
	img, ok := spriteImages[name]
	if !ok {
		return
	}
	// Sprite sheet layout: columns = direction (down=0, up=1, right=2, left=3), rows = frame
	const spriteSize = 32
	sx := dir * spriteSize
	sy := frame * spriteSize
	rect := image.Rect(sx, sy, sx+spriteSize, sy+spriteSize)
	sub := img.SubImage(rect).(*ebt.Image)
	op := &ebt.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	r.screen.DrawImage(sub, op)
}

func (r *EbitenRenderer) DrawText(str string, size int, x, y int, clr app.Color, center bool) {
	face := fontForSize(size)
	if face == nil {
		return
	}
	op := &text.DrawOptions{}
	if center {
		op.PrimaryAlign = text.AlignCenter
	}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(color.RGBA{clr[0], clr[1], clr[2], clr[3]})
	text.Draw(r.screen, str, face, op)
}

func (r *EbitenRenderer) DrawPanel(x, y, w, h int, bg, border app.Color) {
	panel := ebt.NewImage(w, h)
	panel.Fill(color.RGBA{bg[0], bg[1], bg[2], bg[3]})
	op := &ebt.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	r.screen.DrawImage(panel, op)

	borderClr := color.RGBA{border[0], border[1], border[2], border[3]}
	for _, rect := range []struct{ x, y, w, h int }{
		{x, y, w, 2},
		{x, y + h - 2, w, 2},
		{x, y, 2, h},
		{x + w - 2, y, 2, h},
	} {
		b := ebt.NewImage(rect.w, rect.h)
		b.Fill(borderClr)
		bOp := &ebt.DrawImageOptions{}
		bOp.GeoM.Translate(float64(rect.x), float64(rect.y))
		r.screen.DrawImage(b, bOp)
	}
}

func (r *EbitenRenderer) MeasureText(str string, size int) int {
	face := fontForSize(size)
	if face == nil {
		return 0
	}
	w, _ := text.Measure(str, face, 0)
	return int(w) + 1
}

func fontForSize(size int) *text.GoTextFace {
	switch {
	case size <= 12:
		return fontFace12
	case size <= 14:
		return fontFace14
	case size <= 20:
		return fontFace20
	default:
		return fontFace32
	}
}

func hexToRGBA(hex string) color.RGBA {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return color.RGBA{255, 255, 255, 255}
	}
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}
