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
	if input.Options {
		return NewOptionsScene(s, s.audio)
	}
	if input.Credit {
		return NewCreditScene(s)
	}
	if input.Help {
		return &TitleHelpScene{prev: s}
	}
	return nil
}

type TitleHelpScene struct {
	prev Scene
}

func (s *TitleHelpScene) Update(input InputState) Scene {
	if input.Cancel || input.Confirm || input.Help {
		return s.prev
	}
	return nil
}

func (s *TitleHelpScene) Draw(r Renderer) {
	r.Clear(10, 10, 30, 255)
	r.DrawText("操作説明", 20, ScreenWidth/2, 60, Color{255, 220, 100, 255}, true)
	lines := []string{
		"WASD / 矢印キー : 移動・攻撃",
		"X : アクションメニュー",
		"I : アイテム一覧",
		"L : メッセージログ",
		"M : マップ表示",
		"H : ヘルプ",
		"O : オプション",
		"Esc : キャンセル / 戻る",
		"Enter / Space : 決定",
	}
	for i, line := range lines {
		r.DrawText(line, 14, ScreenWidth/2, 120+i*30, Color{200, 200, 200, 255}, true)
	}
	r.DrawText("Esc / H / Enter : 戻る", 12, ScreenWidth/2, ScreenHeight-30, Color{120, 120, 120, 255}, true)
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
