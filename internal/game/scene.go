package game

import "github.com/hajimehoshi/ebiten/v2"

// Scene インターフェース
type Scene interface {
	Update() (Scene, error)
	Draw(screen *ebiten.Image)
}

// App（Ebitenのトップレベル）
type App struct {
	scene Scene
}

func (a *App) Update() error {
	next, err := a.scene.Update()
	if err != nil {
		return err
	}
	if next != nil {
		a.scene = next
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	a.scene.Draw(screen)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
