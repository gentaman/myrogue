package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})

	// マップ描画
	// visible: 明るい色 / explored のみ: 暗い色（静的要素のみ）
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			if !g.explored[x][y] {
				continue
			}
			vis := g.visible[x][y]

			var clr color.Color
			if vis {
				switch g.worldMap[x][y] {
				case Wall:
					clr = color.RGBA{100, 100, 110, 255}
				case Floor:
					clr = color.RGBA{60, 60, 65, 255}
				case Stairs:
					clr = color.RGBA{0, 255, 255, 255}
				case Treasure:
					clr = color.RGBA{255, 215, 0, 255}
				case StairsDown:
					clr = color.RGBA{255, 140, 0, 255}
				case StairsUp:
					clr = color.RGBA{0, 220, 180, 255}
				}
			} else {
				// 記憶中（暗め）
				switch g.worldMap[x][y] {
				case Wall:
					clr = color.RGBA{50, 50, 55, 255}
				case Floor:
					clr = color.RGBA{28, 28, 30, 255}
				case Stairs:
					clr = color.RGBA{0, 120, 120, 255}
				case Treasure:
					clr = color.RGBA{140, 110, 0, 255}
				case StairsDown:
					clr = color.RGBA{140, 70, 0, 255}
				case StairsUp:
					clr = color.RGBA{0, 110, 90, 255}
				}
			}

			rect := ebiten.NewImage(tileSize-1, tileSize-1)
			rect.Fill(clr)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
			screen.DrawImage(rect, op)
		}
	}

	// アイテム描画（explored: 常に表示、visible で色を変える）
	for _, it := range g.mapItems {
		if !g.explored[it.x][it.y] {
			continue
		}
		def := itemDefs[it.kind]
		clr := def.clr
		if !g.visible[it.x][it.y] {
			// 記憶中は暗く
			clr = color.RGBA{clr.R / 2, clr.G / 2, clr.B / 2, 255}
		}
		iRect := ebiten.NewImage(tileSize-5, tileSize-5)
		iRect.Fill(clr)
		iOp := &ebiten.DrawImageOptions{}
		iOp.GeoM.Translate(float64(it.x*tileSize+2), float64(it.y*tileSize+2))
		screen.DrawImage(iRect, iOp)
	}

	// 敵描画（visible のみ）
	for _, e := range g.enemies {
		if !g.visible[e.x][e.y] {
			continue
		}
		def := &enemyDefs[e.kind]
		ox, oy, sz := charAnim(g.frame, e.attackAnim, e.dir)
		eRect := ebiten.NewImage(sz, sz)

		eClr := def.clr
		eRect.Fill(eClr)
		eOp := &ebiten.DrawImageOptions{}
		eOp.GeoM.Translate(float64(e.x*tileSize+ox), float64(e.y*tileSize+oy))
		screen.DrawImage(eRect, eOp)

		// 向きドットの描画（警戒時は赤系で激しく点滅）
		var dotClr color.RGBA
		if e.state == EnemyAlerted {
			// 10フレーム周期で点滅
			if (g.frame/5)%2 == 0 {
				dotClr = color.RGBA{255, 50, 50, 255}
			} else {
				dotClr = color.RGBA{255, 200, 200, 255}
			}
		} else {
			// 通常時は白系でゆるやかに明滅
			alpha := uint8(180 + 40*(g.frame/30%2))
			dotClr = color.RGBA{255, 255, 255, alpha}
		}
		drawDirDot(screen, e.x, e.y, e.dir, dotClr)
	}

	// プレイヤー描画
	playerClr := color.RGBA{255, 100, 100, 255}
	dotClr := color.RGBA{255, 255, 255, 255}
	if g.hasTreasure {
		playerClr = color.RGBA{255, 255, 100, 255}
		dotClr = color.RGBA{64, 64, 64, 255}
	}
	pox, poy, psz := charAnim(g.frame, g.playerAttackAnim, g.playerDir)
	pRect := ebiten.NewImage(psz, psz)
	pRect.Fill(playerClr)
	pOp := &ebiten.DrawImageOptions{}
	pOp.GeoM.Translate(float64(g.playerX*tileSize+pox), float64(g.playerY*tileSize+poy))
	screen.DrawImage(pRect, pOp)
	drawDirDot(screen, g.playerX, g.playerY, g.playerDir, dotClr)

	// アクションメニュー
	if g.menuOpen {
		const (
			rowH      = 32
			padX      = 48 // 左余白（▶ + ラベル開始位置）
			padRight  = 16
			headerH   = 44
			footerH   = 28
			minPanelW = 240
			maxPanelW = screenWidth - 20
		)

		// 動的パネル幅: 最長ラベル幅に合わせる
		hintW := measureText("W/S: 選択  Enter: 決定  X/Esc: 閉じる", fontFace12)
		panelW := minPanelW
		for _, item := range g.menuItems {
			w := measureText(item.label, fontFace14) + padX + padRight
			if w > panelW {
				panelW = w
			}
		}
		if hintW+padRight > panelW {
			panelW = hintW + padRight
		}
		if panelW > maxPanelW {
			panelW = maxPanelW
		}

		// 動的パネル高: 全行収まるか、収まらなければスクロール
		areaH := mapHeight * tileSize
		needH := headerH + len(g.menuItems)*rowH + footerH
		panelH := needH
		if panelH > areaH-8 {
			panelH = areaH - 8
		}
		visibleRows := (panelH - headerH - footerH) / rowH
		if visibleRows < 1 {
			visibleRows = 1
		}

		// スクロールオフセット（カーソルが見える範囲に収める）
		scrollOffset := g.menuCursor - visibleRows + 1
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		if g.menuCursor < scrollOffset {
			scrollOffset = g.menuCursor
		}

		panelX := (screenWidth - panelW) / 2
		panelY := (areaH - panelH) / 2

		drawPanel(screen, panelX, panelY, panelW, panelH, color.RGBA{20, 20, 50, 255}, color.RGBA{100, 150, 255, 255})
		drawText(screen, "アクション", fontFace14, screenWidth/2, panelY+12, color.RGBA{255, 220, 100, 255}, true)

		for i, item := range g.menuItems {
			row := i - scrollOffset
			if row < 0 || row >= visibleRows {
				continue
			}
			y := panelY + headerH + row*rowH
			clr := color.RGBA{180, 180, 180, 255}
			if !item.enabled {
				clr = color.RGBA{80, 80, 80, 255}
			}
			if i == g.menuCursor {
				hl := ebiten.NewImage(panelW-8, 26)
				hl.Fill(color.RGBA{60, 60, 120, 255})
				hlOp := &ebiten.DrawImageOptions{}
				hlOp.GeoM.Translate(float64(panelX+4), float64(y-4))
				screen.DrawImage(hl, hlOp)
				clr = color.RGBA{255, 255, 255, 255}
				if !item.enabled {
					clr = color.RGBA{120, 120, 120, 255}
				}
				drawText(screen, "▶", fontFace12, panelX+14, y, clr, false)
			}
			labelMaxW := panelX + panelW - 16 - (panelX + 30)
			drawText(screen, truncateText(item.label, fontFace14, labelMaxW), fontFace14, panelX+30, y, clr, false)
		}
		drawText(screen, "W/S: 選択  Enter: 決定  X/Esc: 閉じる", fontFace12, screenWidth/2, panelY+panelH-20, color.RGBA{120, 120, 120, 255}, true)
	}

	// タイトルへ戻る確認ダイアログ
	if g.confirmQuit {
		const (
			dW = 300
			dH = 80
			dX = (screenWidth - dW) / 2
			dY = (mapHeight*tileSize - dH) / 2
		)
		drawPanel(screen, dX, dY, dW, dH, color.RGBA{30, 20, 20, 255}, color.RGBA{200, 100, 100, 255})
		drawText(screen, "タイトルへ戻りますか？", fontFace14, screenWidth/2, dY+14, color.RGBA{255, 220, 220, 255}, true)
		drawText(screen, "Enter/Y: はい    Esc/N: いいえ", fontFace12, screenWidth/2, dY+46, color.RGBA{180, 180, 180, 255}, true)
	}

	// ゲームオーバー（勝利・敗北）ダイアログ
	if g.playState == StateWin || g.playState == StateDead {
		const (
			dW = 360
			dH = 120
			dX = (screenWidth - dW) / 2
			dY = (mapHeight*tileSize - dH) / 2
		)
		var pClr, bClr color.RGBA
		var title, sub string
		if g.playState == StateWin {
			pClr = color.RGBA{20, 40, 20, 255}
			bClr = color.RGBA{100, 255, 100, 255}
			title = "--- BEAT THE GAME ---"
			sub = fmt.Sprintf("スコア: %d (ターン: %d / HP: %d)", g.calcScore(), g.turnCount, g.playerHP)
		} else {
			pClr = color.RGBA{40, 20, 20, 255}
			bClr = color.RGBA{255, 100, 100, 255}
			title = "--- GAME OVER ---"
			sub = "あなたはダンジョンで力尽きた..."
		}
		drawPanel(screen, dX, dY, dW, dH, pClr, bClr)
		drawText(screen, title, fontFace14, screenWidth/2, dY+20, bClr, true)
		drawText(screen, sub, fontFace12, screenWidth/2, dY+55, color.RGBA{220, 220, 220, 255}, true)
		drawText(screen, "R: 再挑戦    Esc: タイトルへ", fontFace12, screenWidth/2, dY+dH-25, color.RGBA{180, 180, 180, 255}, true)
	}

	// UIステータスバー（mapHeight*tileSize 以下の80px帯）
	// 行1 (y+0):  フロア/ターン（左）  キーヒント（右）
	// 行2 (y+16): HP  重量（左）
	// 行3 (y+40): メッセージ
	statusY := mapHeight*tileSize + 8

	hpClr := color.RGBA{200, 200, 200, 255}
	if g.playerHP <= playerMaxHP/4 {
		hpClr = color.RGBA{255, 80, 80, 255}
	}
	wt := g.currentWeight()

	drawText(screen, fmt.Sprintf("フロア: %d  ターン: %d", g.floor+1, g.turnCount), fontFace12, 8, statusY, color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, fmt.Sprintf("Lv: %d  XP: %d / %d", g.Level, g.XP, g.XPToNext), fontFace12, 160, statusY, color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, fmt.Sprintf("HP: %d / %d  MP: %d / %d", g.playerHP, playerMaxHP+g.Vit*2, g.MP, g.MaxMP), fontFace12, 8, statusY+16, hpClr, false)
	drawText(screen, fmt.Sprintf("ATK: %d  DEF: %d 重量: %d / %d  アイテム: %d", 1+g.Str+g.equippedAtk(), g.equippedDef()+g.Vit, wt, maxCarryWeight, len(g.inventory)), fontFace12, 160, statusY+16, color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, fmt.Sprintf("Str:%d Wis:%d Fai:%d Vit:%d Agi:%d Luk:%d", g.Str, g.Wis, g.Fai, g.Vit, g.Agi, g.Luk), fontFace12, 8, statusY+32, color.RGBA{150, 150, 150, 255}, false)
	drawText(screen, g.message, fontFace14, 8, statusY+48, color.RGBA{255, 255, 255, 255}, false)

	drawText(screen, "H: ヘルプ", fontFace12, screenWidth-80, statusY, color.RGBA{100, 100, 100, 255}, false)
}
