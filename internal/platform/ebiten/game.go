package ebiten

import (
	ebt "github.com/hajimehoshi/ebiten/v2"

	"github.com/gentaman/myrogue/internal/app"
	"github.com/gentaman/myrogue/internal/core/content"
)

type Game struct {
	app *app.App
}

func NewGame() *Game {
	reg := content.NewRegistry()
	if err := reg.LoadAll(PlayerJSON, EnemiesJSON, CompanionsJSON, ItemsJSON, FloorsJSON); err != nil {
		panic(err)
	}
	audio := &Audio{}
	a := app.NewApp(reg, audio)
	return &Game{app: a}
}

func (g *Game) Update() error {
	input := PollInput()
	g.app.Update(input)
	return nil
}

func (g *Game) Draw(screen *ebt.Image) {
	r := &EbitenRenderer{screen: screen}
	g.app.Draw(r)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func Run() error {
	game := NewGame()
	ebt.SetWindowTitle("My Rogue")
	ebt.SetWindowSize(screenWidth, screenHeight)
	return ebt.RunGame(game)
}
