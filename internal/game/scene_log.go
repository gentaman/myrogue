package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type LogScene struct {
	game         *GameScene
	scrollOffset int
	initialized  bool
}

const (
	logPanelW = 560
	logPanelH = 400
	logPadY   = 40
	logRowH   = 18
)

func (s *LogScene) Update() (Scene, error) {
	logs := s.game.messageLog
	visibleRows := (logPanelH - logPadY - 20) / logRowH

	if !s.initialized {
		s.scrollOffset = len(logs) - visibleRows
		if s.scrollOffset < 0 {
			s.scrollOffset = 0
		}
		s.initialized = true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyL) {
		return s.game, nil
	}

	// スクロール操作
	maxScroll := len(logs) - visibleRows
	if maxScroll < 0 {
		maxScroll = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.scrollOffset--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.scrollOffset++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		s.scrollOffset -= visibleRows
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		s.scrollOffset += visibleRows
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		s.scrollOffset = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		s.scrollOffset = maxScroll
	}

	// 範囲制限
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
	if s.scrollOffset > maxScroll {
		s.scrollOffset = maxScroll
	}

	return nil, nil
}

func (s *LogScene) Draw(screen *ebiten.Image) {
	s.game.Draw(screen) // 背景としてゲーム画面を描画

	// 半透明のオーバーレイ
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	const (
		panelX = (screenWidth - logPanelW) / 2
		panelY = (screenHeight - logPanelH) / 2
		padX   = 20
	)

	drawPanel(screen, panelX, panelY, logPanelW, logPanelH, color.RGBA{20, 20, 30, 255}, color.RGBA{100, 100, 200, 255})
	drawText(screen, "メッセージ履歴", fontFace14, screenWidth/2, panelY+15, color.RGBA{255, 220, 100, 255}, true)

	logs := s.game.messageLog
	visibleRows := (logPanelH - logPadY - 20) / logRowH

	for i := 0; i < visibleRows; i++ {
		idx := s.scrollOffset + i
		if idx >= len(logs) {
			break
		}
		y := panelY + logPadY + i*logRowH
		drawText(screen, logs[idx], fontFace12, panelX+padX, y, color.RGBA{220, 220, 220, 255}, false)
	}

	// スクロールインジケータ
	if len(logs) > visibleRows {
		drawText(screen, fmt.Sprintf("▲ %d / %d ▼", s.scrollOffset+1, len(logs)), fontFace12, screenWidth/2, panelY+logPanelH-20, color.RGBA{150, 150, 150, 255}, true)
	}

	drawText(screen, "W/S: スクロール  L/Esc: 閉じる", fontFace12, screenWidth/2, panelY+logPanelH-5, color.RGBA{100, 100, 100, 255}, true)
}
