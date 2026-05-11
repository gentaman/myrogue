package app

import "fmt"

type OptionsScene struct {
	prev   Scene
	volume int // 0-100
	audio  AudioPlayer
}

func NewOptionsScene(prev Scene, audio AudioPlayer) *OptionsScene {
	return &OptionsScene{prev: prev, volume: 50, audio: audio}
}

func (s *OptionsScene) Update(input InputState) Scene {
	if input.Cancel || input.Options {
		return s.prev
	}
	if input.Left {
		s.volume -= 10
		if s.volume < 0 {
			s.volume = 0
		}
	}
	if input.Right {
		s.volume += 10
		if s.volume > 100 {
			s.volume = 100
		}
	}
	return nil
}

func (s *OptionsScene) Draw(r Renderer) {
	r.Clear(10, 10, 30, 255)
	r.DrawText("オプション", 20, ScreenWidth/2, 80, Color{255, 220, 100, 255}, true)

	r.DrawText("SE 音量", 14, ScreenWidth/2, 180, Color{220, 220, 220, 255}, true)

	const (
		barX = 160
		barY = 220
		barW = 320
		barH = 12
	)
	r.DrawRect(barX, barY, barW, barH, Color{60, 60, 80, 255})

	filled := s.volume * barW / 100
	if filled > 0 {
		r.DrawRect(barX, barY, filled, barH, Color{100, 200, 255, 255})
	}

	knobX := barX + filled - 9
	r.DrawRect(knobX, barY-3, 18, 18, Color{255, 255, 255, 255})

	r.DrawText(fmt.Sprintf("%d%%", s.volume), 14, ScreenWidth/2, barY+40, Color{255, 255, 255, 255}, true)
	r.DrawText("← / → : 音量調整", 12, ScreenWidth/2, 320, Color{180, 180, 180, 255}, true)
	r.DrawText("Esc / O : 戻る", 12, ScreenWidth/2, 345, Color{150, 150, 150, 255}, true)
}
