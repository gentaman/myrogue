package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// タイトル画面
type TitleScene struct{}

func (s *TitleScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return NewGameScene(), nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		return NewCharacterCreateScene(), nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return &HelpScene{prev: s}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		return &OptionsScene{prev: s}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		return NewCreditScene(s), nil
	}
	return nil, nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "My Rogue", fontFace32, screenWidth/2, 120, color.RGBA{255, 220, 100, 255}, true)
	drawText(screen, "ダンジョンを探索し、宝を持ち帰れ", fontFace14, screenWidth/2, 180, color.RGBA{200, 200, 200, 255}, true)

	drawText(screen, "Enter / Space : ゲーム開始", fontFace14, screenWidth/2, 280, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "E : キャラクター作成", fontFace14, screenWidth/2, 310, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "H : 操作説明", fontFace14, screenWidth/2, 340, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "O : オプション", fontFace14, screenWidth/2, 370, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "C : クレジット", fontFace14, screenWidth/2, 400, color.RGBA{255, 255, 255, 255}, true)
}
