package game

import "github.com/hajimehoshi/ebiten/v2"

// Run はゲームを起動するエントリポイント
func Run() error {
	app := &App{scene: &TitleScene{}}
	ebiten.SetWindowTitle("My Rogue")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	return ebiten.RunGame(app)
}
