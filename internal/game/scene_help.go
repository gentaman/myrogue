package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ヘルプ画面
type HelpScene struct {
	prev Scene // 戻り先（タイトルまたはゲーム画面）
}

func (s *HelpScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return s.prev, nil
	}
	return nil, nil
}

func (s *HelpScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "操作説明", fontFace20, screenWidth/2, 60, color.RGBA{255, 220, 100, 255}, true)

	lines := []string{
		"移動: 矢印キー / WASD",
		"攻撃: 敵のいる方向に移動（1ターン消費）",
		"",
		"目的: 宝（金色）を見つけて入り口（水色）に戻る",
		"    敵を避けるか倒しながら生き延びろ！",
		"",
		"色の意味:",
		"  灰色 = 壁    暗い灰 = 床",
		"  水色 = 階段（入り口）",
		"  金色 = 宝    紫色 = 敵",
		"  赤色 = プレイヤー",
		"  黄色 = プレイヤー（宝所持）",
		"",
		"HP: 20からスタート。敵に隣接されると攻撃される",
		"R: ゲームオーバー/クリア後にリスタート",
		"Esc: タイトルに戻る",
	}

	y := 120
	for _, line := range lines {
		if line == "" {
			y += 10
			continue
		}
		drawText(screen, line, fontFace14, 80, y, color.RGBA{220, 220, 220, 255}, false)
		y += 24
	}

	drawText(screen, "Esc / Backspace / H : 戻る", fontFace12, screenWidth/2, 430, color.RGBA{150, 150, 150, 255}, true)
}
