package app

type HelpScene struct {
	game *GameScene
}

func (s *HelpScene) Update(input InputState) Scene {
	if input.Cancel || input.Confirm || input.Help {
		return s.game
	}
	return nil
}

func (s *HelpScene) Draw(r Renderer) {
	s.game.Draw(r)

	const (
		panelW = 400
		panelH = 340
	)
	panelX := (ScreenWidth - panelW) / 2
	panelY := (ScreenHeight - panelH) / 2
	r.DrawPanel(panelX, panelY, panelW, panelH, Color{20, 20, 40, 255}, Color{100, 150, 255, 255})
	r.DrawText("操作説明", 14, ScreenWidth/2, panelY+12, Color{255, 220, 100, 255}, true)

	lines := []string{
		"WASD / 矢印キー : 移動・攻撃",
		"X : メニューを開く",
		"I : アイテム",
		"Z : スキル",
		"L : メッセージログ",
		"M : マップ表示",
		"H : この画面",
		"Esc : 戻る / タイトルに戻る",
		"Enter/Space : 決定",
		"F1 : デバッグHUD",
	}
	for i, line := range lines {
		r.DrawText(line, 12, panelX+24, panelY+48+i*28, Color{200, 200, 200, 255}, false)
	}
	r.DrawText("Esc/H/Enter: 閉じる", 12, ScreenWidth/2, panelY+panelH-24, Color{120, 120, 120, 255}, true)
}
