package app

import "github.com/gentaman/myrogue/internal/core/content"

type TitleScene struct {
	registry *content.Registry
	audio    AudioPlayer
}

func NewTitleScene() *TitleScene {
	return &TitleScene{}
}

func NewTitleSceneWithDeps(reg *content.Registry, audio AudioPlayer) *TitleScene {
	return &TitleScene{registry: reg, audio: audio}
}

func (s *TitleScene) Update(input InputState) Scene {
	if input.Confirm {
		if s.registry != nil {
			return NewGameScene(s.registry, s.audio)
		}
	}
	return nil
}

func (s *TitleScene) Draw(r Renderer) {
	r.Clear(10, 10, 30, 255)
	r.DrawText("My Rogue", 32, ScreenWidth/2, 120, Color{255, 220, 100, 255}, true)
	r.DrawText("ダンジョンを探索し、宝を持ち帰れ", 14, ScreenWidth/2, 180, Color{200, 200, 200, 255}, true)
	r.DrawText("Enter / Space : ゲーム開始", 14, ScreenWidth/2, 280, Color{255, 255, 255, 255}, true)
	r.DrawText("E : キャラクター作成", 14, ScreenWidth/2, 310, Color{255, 255, 255, 255}, true)
	r.DrawText("H : 操作説明", 14, ScreenWidth/2, 340, Color{255, 255, 255, 255}, true)
	r.DrawText("O : オプション", 14, ScreenWidth/2, 370, Color{255, 255, 255, 255}, true)
	r.DrawText("C : クレジット", 14, ScreenWidth/2, 400, Color{255, 255, 255, 255}, true)
}
