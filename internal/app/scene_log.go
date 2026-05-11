package app

type LogScene struct {
	game         *GameScene
	scrollOffset int
}

func (s *LogScene) Update(input InputState) Scene {
	if input.Cancel || input.Log {
		return s.game
	}
	if input.Up {
		s.scrollOffset--
		if s.scrollOffset < 0 {
			s.scrollOffset = 0
		}
	}
	if input.Down {
		maxScroll := len(s.game.MessageLog) - 10
		if maxScroll < 0 {
			maxScroll = 0
		}
		s.scrollOffset++
		if s.scrollOffset > maxScroll {
			s.scrollOffset = maxScroll
		}
	}
	return nil
}

func (s *LogScene) Draw(r Renderer) {
	s.game.Draw(r)

	const (
		panelW     = 500
		panelH     = 320
		lineH      = 20
		headerH    = 40
		footerH    = 24
		maxVisible = 12
	)
	panelX := (ScreenWidth - panelW) / 2
	panelY := (ScreenHeight - panelH) / 2
	r.DrawPanel(panelX, panelY, panelW, panelH, Color{15, 15, 30, 255}, Color{80, 120, 200, 255})
	r.DrawText("メッセージログ", 14, ScreenWidth/2, panelY+12, Color{200, 220, 255, 255}, true)

	if s.scrollOffset == 0 {
		maxScroll := len(s.game.MessageLog) - maxVisible
		if maxScroll > 0 {
			s.scrollOffset = maxScroll
		}
	}

	for i := 0; i < maxVisible; i++ {
		idx := s.scrollOffset + i
		if idx < 0 || idx >= len(s.game.MessageLog) {
			continue
		}
		y := panelY + headerH + i*lineH
		r.DrawText(s.game.MessageLog[idx], 12, panelX+16, y, Color{180, 180, 180, 255}, false)
	}
	r.DrawText("W/S: スクロール  L/Esc: 閉じる", 12, ScreenWidth/2, panelY+panelH-16, Color{100, 100, 100, 255}, true)
}
