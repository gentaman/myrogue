package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type CompanionMenuScene struct {
	game         *GameScene
	companionIdx int
	cursor       int
}

func (s *CompanionMenuScene) Update() (Scene, error) {
	c := &s.game.companions[s.companionIdx]

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return s.game, nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.cursor--
		if s.cursor < 0 {
			s.cursor = 4 // インベントリ, ついてきて, 待ってて, 積極的に, キャンセル
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.cursor++
		if s.cursor > 4 {
			s.cursor = 0
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch s.cursor {
		case 0: // インベントリ
			return &InventoryScene{game: s.game, target: &c.Actor}, nil
		case 1: // ついてきて
			c.order = OrderFollow
			s.game.pushMessage(fmt.Sprintf("%sに「ついてきて」と命令した。", c.GetName()))
			return s.game, nil
		case 2: // 待ってて
			c.order = OrderWait
			s.game.pushMessage(fmt.Sprintf("%sに「待ってて」と命令した。", c.GetName()))
			return s.game, nil
		case 3: // 積極的に
			c.order = OrderAggressive
			s.game.pushMessage(fmt.Sprintf("%sに「積極的に戦って」と命令した。", c.GetName()))
			return s.game, nil
		case 4: // キャンセル
			return s.game, nil
		}
	}

	return nil, nil
}

func (s *CompanionMenuScene) Draw(screen *ebiten.Image) {
	s.game.Draw(screen)

	// 半透明オーバーレイ
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 150})
	screen.DrawImage(overlay, nil)

	c := &s.game.companions[s.companionIdx]
	title := fmt.Sprintf("%sへの指示 (現在の命令: %s)", c.GetName(), CompanionOrderNames[c.order])

	options := []string{
		"持ち物を見る",
		"ついてきて (デフォルト)",
		"待ってて (その場に留まる)",
		"積極的に戦って (中立も攻撃)",
		"やめる",
	}

	panelW, panelH := 300, 200
	px, py := (screenWidth-panelW)/2, (screenHeight-panelH)/2
	drawPanel(screen, px, py, panelW, panelH, color.RGBA{30, 30, 40, 255}, color.RGBA{200, 200, 200, 255})

	drawText(screen, title, fontFace14, px+panelW/2, py+20, color.White, true)

	for i, opt := range options {
		clr := color.RGBA{200, 200, 200, 255}
		if i == s.cursor {
			clr = color.RGBA{255, 255, 100, 255}
			drawText(screen, ">", fontFace14, px+30, py+60+i*25, clr, false)
		}
		drawText(screen, opt, fontFace14, px+50, py+60+i*25, clr, false)
	}
}
