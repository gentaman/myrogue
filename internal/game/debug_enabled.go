//go:build debug

package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *GameScene) drawDebug(screen *ebiten.Image) {
	const (
		panelW = 240
		panelH = 200
		padX   = 10
		rowH   = 16
	)

	x := screenWidth - panelW - 10
	y := 10

	drawPanel(screen, x, y, panelW, panelH, color.RGBA{0, 0, 0, 200}, color.RGBA{255, 0, 0, 255})
	drawText(screen, "DEBUG INFO", fontFace12, x+panelW/2, y+10, color.RGBA{255, 50, 50, 255}, true)

	curY := y + 25

	// フレームとターン情報
	drawText(screen, fmt.Sprintf("Frame: %d", g.frame), fontFace12, x+padX, curY, color.White, false)
	curY += rowH
	stateStr := "PlayerInput"
	if g.turnState == TurnEnemyAct {
		stateStr = fmt.Sprintf("EnemyAct (%d/%d)", g.activeEnemyIdx, len(g.enemies))
	}
	drawText(screen, fmt.Sprintf("State: %s", stateStr), fontFace12, x+padX, curY, color.White, false)
	curY += rowH

	// アニメーション
	animStr := fmt.Sprintf("P_Atk:%d P_Dmg:%d", g.Player.AttackAnim, g.Player.DamageAnim)
	drawText(screen, animStr, fontFace12, x+padX, curY, color.RGBA{255, 200, 100, 255}, false)
	curY += rowH

	// プレイヤー座標
	drawText(screen, fmt.Sprintf("Player: (%d, %d)", g.Player.X, g.Player.Y), fontFace12, x+padX, curY, color.RGBA{0, 255, 255, 255}, false)
	curY += rowH

	// アクティブな敵または近くの敵
	if len(g.enemies) > 0 {
		curY += 5
		drawText(screen, "Enemies:", fontFace12, x+padX, curY, color.RGBA{150, 150, 255, 255}, false)
		curY += rowH

		count := 0
		for i, e := range g.enemies {
			if count >= 5 {
				drawText(screen, "...", fontFace12, x+padX, curY, color.RGBA{150, 150, 150, 255}, false)
				break
			}
			activeMarker := "  "
			if g.turnState == TurnEnemyAct && i == g.activeEnemyIdx {
				activeMarker = "> "
			}
			eStr := fmt.Sprintf("%s%d: (%d,%d) HP:%d", activeMarker, i, e.X, e.Y, e.HP)
			var clr color.Color = color.White
			if e.DamageAnim > 0 {
				clr = color.RGBA{255, 100, 100, 255}
			}
			drawText(screen, eStr, fontFace12, x+padX, curY, clr, false)
			curY += rowH
			count++
		}
	}
}
