package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ヘルプ画面
type HelpScene struct {
	prev         Scene // 戻り先（タイトルまたはゲーム画面）
	scrollOffset int
	content      []helpLine
}

type helpLine struct {
	text string
	clr  color.Color
	size int // 0: Small(12), 1: Normal(14), 2: Large(20)
}

func (s *HelpScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return s.prev, nil
	}

	// スクロール
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 {
		s.scrollOffset -= 2
	} else if wheelY < 0 {
		s.scrollOffset += 2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		s.scrollOffset--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		s.scrollOffset++
	}

	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
	// 最大スクロール量は Draw で計算される内容の高さに依存

	return nil, nil
}

func (s *HelpScene) buildContent() {
	s.content = []helpLine{
		{"操作説明", color.RGBA{255, 220, 100, 255}, 2},
		{"", color.White, 1},
		{"【基本操作】", color.RGBA{100, 255, 100, 255}, 1},
		{"移動 / 攻撃: 矢印キー or WASD", color.White, 1},
		{"待機 (1ターン): スペースキー", color.White, 1},
		{"メニュー開閉: X キー", color.White, 1},
		{"", color.White, 1},
		{"【画面切替】", color.RGBA{100, 255, 100, 255}, 1},
		{"インベントリ: I キー", color.White, 1},
		{"メッセージ履歴: L キー", color.White, 1},
		{"全体マップ: M キー", color.White, 1},
		{"オプション: O キー", color.White, 1},
		{"ヘルプ: H キー", color.White, 1},
		{"", color.White, 1},
		{"【システム】", color.RGBA{100, 255, 100, 255}, 1},
		{"インベントリ: 武器や防具を装備することでステータスが上昇します。", color.RGBA{200, 200, 200, 255}, 1},
		{"耐久度: 装備品は使用するごとに劣化し、0になると壊れます。", color.RGBA{200, 200, 200, 255}, 1},
		{"重量: 持ち物の重量が最大運搬量を超えると移動できなくなります。", color.RGBA{200, 200, 200, 255}, 1},
		{"感情値: 他の種族に攻撃すると敵対します。同じ種族とは友好的です。", color.RGBA{200, 200, 200, 255}, 1},
		{"レベルアップ: 敵を倒すと経験値を得て、ステータスが上昇します。", color.RGBA{200, 200, 200, 255}, 1},
		{"", color.White, 1},
		{"【シンボル】", color.RGBA{100, 255, 100, 255}, 1},
		{"@ (赤/黄): あなた (黄は宝所持)", color.RGBA{255, 100, 100, 255}, 1},
		{"# (灰色): 壁 (通行不可)", color.RGBA{150, 150, 150, 255}, 1},
		{". (暗灰): 床", color.RGBA{100, 100, 100, 255}, 1},
		{"< (水色): 上り階段 (帰還/前の階へ)", color.RGBA{100, 200, 255, 255}, 1},
		{"> (水色): 下り階段 (次の階へ)", color.RGBA{100, 200, 255, 255}, 1},
		{"* (金色): 宝 / 宝箱", color.RGBA{255, 215, 0, 255}, 1},
		{"A~Z (紫色): 敵ユニット", color.RGBA{200, 100, 255, 255}, 1},
		{"", color.White, 1},
	}

	// デバッグ用情報の追加
	s.addDebugHelp()

	s.content = append(s.content, helpLine{"", color.White, 1})
	s.content = append(s.content, helpLine{"Esc / Backspace / H : 戻る", color.RGBA{150, 150, 150, 255}, 0})
}

func (s *HelpScene) Draw(screen *ebiten.Image) {
	if len(s.content) == 0 {
		s.buildContent()
	}

	screen.Fill(color.RGBA{10, 10, 30, 255})

	lineH := 24
	startY := 60

	// スクロール限界調整
	maxScroll := len(s.content) - (screenHeight-startY-40)/lineH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.scrollOffset > maxScroll {
		s.scrollOffset = maxScroll
	}

	for i, line := range s.content {
		drawIdx := i - s.scrollOffset
		y := startY + drawIdx*lineH

		if y < 40 || y > screenHeight-20 {
			continue
		}

		face := fontFace14
		isCenter := false
		if line.size == 2 {
			face = fontFace20
			isCenter = true
		} else if line.size == 0 {
			face = fontFace12
			isCenter = true
		}

		x := 60
		if isCenter {
			x = screenWidth / 2
		}

		if line.text != "" {
			drawText(screen, line.text, face, x, y, line.clr, isCenter)
		}
	}

	// スクロールガイド（上）
	if s.scrollOffset > 0 {
		drawText(screen, "▲ スクロール ▲", fontFace12, screenWidth/2, 40, color.RGBA{100, 100, 100, 255}, true)
	}
	// スクロールガイド（下）
	if len(s.content)*lineH > screenHeight-startY-40 {
		if s.scrollOffset < maxScroll {
			drawText(screen, "▼ スクロール ▼", fontFace12, screenWidth/2, screenHeight-20, color.RGBA{100, 100, 100, 255}, true)
		}
	}
}
