package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// オプション画面
type OptionsScene struct {
	prev Scene
}

func (s *OptionsScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyO) {
		return s.prev, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		sfxVolume -= 0.1
		if sfxVolume < 0 {
			sfxVolume = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		sfxVolume += 0.1
		if sfxVolume > 1.0 {
			sfxVolume = 1.0
		}
	}
	return nil, nil
}

func (s *OptionsScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "オプション", fontFace20, screenWidth/2, 80, color.RGBA{255, 220, 100, 255}, true)

	// SE音量スライダー
	const (
		barX     = 160
		barY     = 220
		barW     = 320
		barH     = 12
		knobSize = 18
	)
	drawText(screen, "SE 音量", fontFace14, screenWidth/2, 180, color.RGBA{220, 220, 220, 255}, true)

	// バー背景
	bar := ebiten.NewImage(barW, barH)
	bar.Fill(color.RGBA{60, 60, 80, 255})
	barOp := &ebiten.DrawImageOptions{}
	barOp.GeoM.Translate(float64(barX), float64(barY))
	screen.DrawImage(bar, barOp)

	// バー塗り
	filled := int(sfxVolume * float64(barW))
	if filled > 0 {
		fill := ebiten.NewImage(filled, barH)
		fill.Fill(color.RGBA{100, 200, 255, 255})
		fillOp := &ebiten.DrawImageOptions{}
		fillOp.GeoM.Translate(float64(barX), float64(barY))
		screen.DrawImage(fill, fillOp)
	}

	// ノブ
	knobX := barX + filled - knobSize/2
	knob := ebiten.NewImage(knobSize, knobSize)
	knob.Fill(color.RGBA{255, 255, 255, 255})
	knobOp := &ebiten.DrawImageOptions{}
	knobOp.GeoM.Translate(float64(knobX), float64(barY)-float64(knobSize-barH)/2)
	screen.DrawImage(knob, knobOp)

	pct := int(sfxVolume * 100)
	drawText(screen, fmt.Sprintf("%d%%", pct), fontFace14, screenWidth/2, barY+40, color.RGBA{255, 255, 255, 255}, true)

	drawText(screen, "<- / -> : 音量調整", fontFace12, screenWidth/2, 320, color.RGBA{180, 180, 180, 255}, true)
	drawText(screen, "Esc / O : 戻る", fontFace12, screenWidth/2, 345, color.RGBA{150, 150, 150, 255}, true)
}
