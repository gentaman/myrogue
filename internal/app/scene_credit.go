package app

import "strings"

const creditLineHeight = 28

type CreditScene struct {
	prev    Scene
	lines   []string
	scrollY float64
}

func NewCreditScene(prev Scene) *CreditScene {
	text := `My Rogue

---

FONT

M+ FONTS (mplus1.ttf)
Copyright 2021 The M+ FONTS Project Authors
License: SIL Open Font License 1.1

---

SOUND EFFECTS

8-bit Sound Effects Pack 001
@Shades

---

GAME ENGINE

Ebiten v2
Copyright 2013 Hajime Hoshi
License: Apache License 2.0

---

PROGRAMMING

gentaman

---

Special Thanks

You

---`
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	return &CreditScene{prev: prev, lines: lines}
}

func (s *CreditScene) Update(input InputState) Scene {
	if input.Cancel || input.Credit {
		return s.prev
	}
	s.scrollY += 1.0
	totalH := float64(len(s.lines) * creditLineHeight)
	if s.scrollY > totalH+float64(ScreenHeight) {
		s.scrollY = 0
	}
	return nil
}

func (s *CreditScene) Draw(r Renderer) {
	r.Clear(5, 5, 20, 255)

	startY := float64(ScreenHeight) - s.scrollY
	for i, line := range s.lines {
		y := startY + float64(i*creditLineHeight)
		if y < -float64(creditLineHeight) || y > float64(ScreenHeight) {
			continue
		}
		if line == "---" {
			r.DrawRect(ScreenWidth/2-200, int(y)+creditLineHeight/2, 400, 1, Color{80, 80, 120, 255})
			continue
		}
		clr := Color{200, 200, 200, 255}
		size := 14
		if i == 0 {
			size = 20
			clr = Color{255, 220, 100, 255}
		}
		r.DrawText(line, size, ScreenWidth/2, int(y), clr, true)
	}

	r.DrawText("Esc / C : 戻る", 12, ScreenWidth/2, ScreenHeight-20, Color{100, 100, 100, 255}, true)
}
