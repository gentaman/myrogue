package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})

	// マップ描画
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			if !g.explored[x][y] {
				continue
			}

			var clr color.Color
			switch g.worldMap[x][y] {
			case Wall:
				clr = color.RGBA{80, 80, 80, 255}
			case Floor:
				clr = color.RGBA{40, 40, 40, 255}
			case Stairs:
				clr = color.RGBA{0, 255, 255, 255}
			case Treasure:
				clr = color.RGBA{255, 215, 0, 255}
			case StairsDown:
				clr = color.RGBA{255, 140, 0, 255} // オレンジ（下り）
			case StairsUp:
				clr = color.RGBA{0, 220, 180, 255} // 緑がかった水色（上り）
			}

			rect := ebiten.NewImage(tileSize-1, tileSize-1)
			rect.Fill(clr)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
			screen.DrawImage(rect, op)
		}
	}

	// 敵描画（探索済みエリアのみ）
	for _, e := range g.enemies {
		if g.explored[e.x][e.y] {
			eRect := ebiten.NewImage(tileSize-1, tileSize-1)
			eRect.Fill(color.RGBA{200, 50, 200, 255})
			eOp := &ebiten.DrawImageOptions{}
			eOp.GeoM.Translate(float64(e.x*tileSize), float64(e.y*tileSize))
			screen.DrawImage(eRect, eOp)
			drawDirDot(screen, e.x, e.y, e.dir, color.RGBA{255, 200, 255, 255})
		}
	}

	// プレイヤー描画
	playerClr := color.RGBA{255, 100, 100, 255}
	if g.hasTreasure {
		playerClr = color.RGBA{255, 255, 100, 255}
	}
	pRect := ebiten.NewImage(tileSize-1, tileSize-1)
	pRect.Fill(playerClr)
	pOp := &ebiten.DrawImageOptions{}
	pOp.GeoM.Translate(float64(g.playerX*tileSize), float64(g.playerY*tileSize))
	screen.DrawImage(pRect, pOp)
	drawDirDot(screen, g.playerX, g.playerY, g.playerDir, color.RGBA{255, 255, 255, 255})

	// アクションメニュー
	if g.menuOpen {
		const (
			panelW = 320
			panelH = 160
			panelX = (screenWidth - panelW) / 2
			panelY = (mapHeight*tileSize - panelH) / 2
		)
		// 背景パネル
		panel := ebiten.NewImage(panelW, panelH)
		panel.Fill(color.RGBA{20, 20, 50, 230})
		panelOp := &ebiten.DrawImageOptions{}
		panelOp.GeoM.Translate(float64(panelX), float64(panelY))
		screen.DrawImage(panel, panelOp)
		// 枠線
		for _, r := range []struct{ x, y, w, h int }{
			{panelX, panelY, panelW, 2},
			{panelX, panelY + panelH - 2, panelW, 2},
			{panelX, panelY, 2, panelH},
			{panelX + panelW - 2, panelY, 2, panelH},
		} {
			border := ebiten.NewImage(r.w, r.h)
			border.Fill(color.RGBA{100, 150, 255, 255})
			bOp := &ebiten.DrawImageOptions{}
			bOp.GeoM.Translate(float64(r.x), float64(r.y))
			screen.DrawImage(border, bOp)
		}

		drawText(screen, "アクション", fontFace14, screenWidth/2, panelY+12, color.RGBA{255, 220, 100, 255}, true)

		for i, item := range g.menuItems {
			y := panelY + 44 + i*32
			clr := color.RGBA{180, 180, 180, 255}
			if !item.enabled {
				clr = color.RGBA{80, 80, 80, 255}
			}
			if i == g.menuCursor {
				// 選択行ハイライト
				hl := ebiten.NewImage(panelW-8, 26)
				hl.Fill(color.RGBA{60, 60, 120, 200})
				hlOp := &ebiten.DrawImageOptions{}
				hlOp.GeoM.Translate(float64(panelX+4), float64(y-4))
				screen.DrawImage(hl, hlOp)
				clr = color.RGBA{255, 255, 255, 255}
				if !item.enabled {
					clr = color.RGBA{120, 120, 120, 255}
				}
				drawText(screen, "▶", fontFace12, panelX+14, y, clr, false)
			}
			drawText(screen, item.label, fontFace14, panelX+30, y, clr, false)
		}
		drawText(screen, "W/S: 選択  Enter: 決定  X/Esc: 閉じる", fontFace12, screenWidth/2, panelY+panelH-20, color.RGBA{120, 120, 120, 255}, true)
	}

	// UIメッセージ
	msgY := mapHeight*tileSize + 8
	drawText(screen, fmt.Sprintf("フロア: %d  ターン: %d  HP: %d", g.floor+1, g.turnCount, g.playerHP), fontFace12, 8, msgY, color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, g.message, fontFace14, 8, msgY+20, color.RGBA{255, 255, 255, 255}, false)

	if g.playState == StateWin {
		drawText(screen, "R: 再挑戦 / Esc: タイトルへ", fontFace12, 8, msgY+44, color.RGBA{150, 255, 150, 255}, false)
	} else if g.playState == StateDead {
		drawText(screen, "R: 再挑戦 / Esc: タイトルへ", fontFace12, 8, msgY+44, color.RGBA{255, 100, 100, 255}, false)
	} else {
		drawText(screen, "X: アクション / H: ヘルプ / O: オプション / Esc: タイトル", fontFace12, screenWidth-400, msgY, color.RGBA{100, 100, 100, 255}, false)
	}
}
