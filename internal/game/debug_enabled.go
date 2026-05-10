//go:build debug

package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *GameScene) handleDebugKeys() (Scene, error) {
	// アニメーション速度調整
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) { // ]
		g.AnimSpeed += 0.5
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) { // [
		if g.AnimSpeed > 0.5 {
			g.AnimSpeed -= 0.5
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyP) { // Reset Speed
		g.AnimSpeed = 1.0
	}

	// 詳細デバッグ画面へ
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		ds := &DebugScene{game: g}
		ds.refresh()
		return ds, nil
	}

	return nil, nil
}

func (g *GameScene) drawDebug(screen *ebiten.Image) {
	const (
		panelW = 280
		panelH = 400
		padX   = 10
		rowH   = 14
	)

	x := screenWidth - panelW - 10
	y := 10

	drawPanel(screen, x, y, panelW, panelH, color.RGBA{0, 0, 0, 220}, color.RGBA{255, 0, 0, 255})
	drawText(screen, "--- DEBUG HUD ---", fontFace12, x+panelW/2, y+10, color.RGBA{255, 100, 100, 255}, true)

	curY := y + 25

	// 基本情報
	drawText(screen, fmt.Sprintf("Frame: %d  Speed: %.1fx", g.frame, g.AnimSpeed), fontFace12, x+padX, curY, color.White, false)
	curY += rowH
	stateStr := "PlayerInput"
	if g.turnState == TurnEnemyAct {
		stateStr = fmt.Sprintf("EnemyAct (%d/%d)", g.activeEnemyIdx, len(g.enemies))
	}
	drawText(screen, fmt.Sprintf("State: %s", stateStr), fontFace12, x+padX, curY, color.White, false)
	curY += rowH
	curY += 4

	// プレイヤー詳細
	drawText(screen, "[Player Status]", fontFace12, x+padX, curY, color.RGBA{100, 255, 100, 255}, false)
	curY += rowH
	drawText(screen, fmt.Sprintf("ID: %d  Race: %s", g.Player.ID, RaceNames[g.Player.Race]), fontFace12, x+padX+4, curY, color.White, false)
	curY += rowH
	drawText(screen, fmt.Sprintf("Pos: (%d, %d)  HP: %d/%d", g.Player.X, g.Player.Y, g.Player.HP, g.Player.MaxHP), fontFace12, x+padX+4, curY, color.White, false)
	curY += rowH
	drawText(screen, fmt.Sprintf("Str:%d Wis:%d Fai:%d Vit:%d Agi:%d Luk:%d", g.Player.Str, g.Player.Wis, g.Player.Fai, g.Player.Vit, g.Player.Agi, g.Player.Luk), fontFace12, x+padX+4, curY, color.White, false)
	curY += rowH
	curY += 6

	// エネミーリスト
	drawText(screen, fmt.Sprintf("[Enemies: %d]", len(g.enemies)), fontFace12, x+padX, curY, color.RGBA{100, 100, 255, 255}, false)
	curY += rowH

	count := 0
	for i := range g.enemies {
		if count >= 15 {
			drawText(screen, fmt.Sprintf("...and %d more", len(g.enemies)-count), fontFace12, x+padX+4, curY, color.Gray16{0x8888}, false)
			break
		}
		e := &g.enemies[i]

		activeMarker := "  "
		if g.turnState == TurnEnemyAct && i == g.activeEnemyIdx {
			activeMarker = "> "
		}

		// ID, 名前, Lv, HP, 性格, プレイヤーへの感情
		rel := e.Relations[g.Player.ID]
		eStr := fmt.Sprintf("%s%d:%s Lv%d HP%d %s", activeMarker, i, enemyDefs[e.kind].Name, e.Level, e.HP, RaceNames[e.Race])
		drawText(screen, eStr, fontFace12, x+padX+4, curY, color.White, false)
		curY += rowH

		// 追加情報
		relStr := fmt.Sprintf("    Rel:%d Pos:(%d,%d)", rel, e.X, e.Y)
		drawText(screen, relStr, fontFace12, x+padX+4, curY, color.RGBA{200, 200, 200, 255}, false)
		curY += rowH

		count++
	}

	drawText(screen, "F1: Full Debug | [: Speed-  ]: Speed+  P: Reset", fontFace12, x+panelW/2, y+panelH-10, color.RGBA{150, 150, 150, 255}, true)
}
