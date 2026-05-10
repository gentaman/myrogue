//go:build !debug

package game

import "github.com/hajimehoshi/ebiten/v2"

func (g *GameScene) drawDebug(screen *ebiten.Image) {}

func (g *GameScene) handleDebugKeys() (Scene, error) {
	return nil, nil
}
