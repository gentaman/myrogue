package game

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func (g *GameScene) Draw(screen *ebiten.Image) {
	sw, sh := screen.Size()

	// 1. ゲームプレイエリア（上部 5/6）
	upperRect := image.Rect(0, 0, sw, gameplayAreaHeight)
	upperScreen := screen.SubImage(upperRect).(*ebiten.Image)

	// 2. UIエリア（下部 1/6）
	lowerRect := image.Rect(0, gameplayAreaHeight, sw, sh)
	lowerScreen := screen.SubImage(lowerRect).(*ebiten.Image)

	// それぞれの領域に対して描画
	g.drawWorld(upperScreen)
	g.drawUI(lowerScreen, float64(gameplayAreaHeight))

	// 3. ダイアログ・メニュー（全画面オーバーレイ）
	g.drawOverlays(screen)
}

func (g *GameScene) drawWorld(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 20, 255})
	camX, camY := g.cameraOffset()

	// マップ描画
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
				case StairsDown:
					clr = color.RGBA{255, 140, 0, 255}
				case StairsUp:
					clr = color.RGBA{0, 220, 180, 255}
				}
			} else {
				switch g.worldMap[x][y] {
				case Wall:
					clr = color.RGBA{50, 50, 55, 255}
				case Floor:
					clr = color.RGBA{28, 28, 30, 255}
				case Stairs:
					clr = color.RGBA{0, 120, 120, 255}
				case StairsDown:
					clr = color.RGBA{140, 70, 0, 255}
				case StairsUp:
					clr = color.RGBA{0, 110, 90, 255}
				}
			}
			rect := ebiten.NewImage(tileSize-2, tileSize-2)
			rect.Fill(clr)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tileSize+1)-camX, float64(y*tileSize+1)-camY)
			screen.DrawImage(rect, op)
		}
	}

	// アイテム・宝箱描画
	for _, it := range g.mapItems {
		if !g.explored[it.X][it.Y] {
			continue
		}

		var clr color.RGBA
		if len(it.Inventory) > 1 {
			// 複数あれば宝箱（金・茶）
			clr = color.RGBA{180, 130, 40, 255}
		} else if len(it.Inventory) == 1 {
			// 1つならそのアイテムの色
			clr = itemDefs[it.Inventory[0].kind].clr
		} else {
			continue
		}

		if !g.visible[it.X][it.Y] {
			clr = color.RGBA{clr.R / 2, clr.G / 2, clr.B / 2, 255}
		}
		iRect := ebiten.NewImage(tileSize-16, tileSize-16)
		iRect.Fill(clr)
		iOp := &ebiten.DrawImageOptions{}
		iOp.GeoM.Translate(float64(it.X*tileSize+8)-camX, float64(it.Y*tileSize+8)-camY)
		screen.DrawImage(iRect, iOp)
	}

	// 敵描画
	for _, e := range g.enemies {
		if !g.visible[e.X][e.Y] {
			continue
		}
		def := &enemyDefs[e.kind]
		g.drawActor(screen, &e.Actor, nil, def.Color, camX, camY)
		var dotClr color.RGBA
		if e.state == EnemyAlerted {
			if (g.frame/5)%2 == 0 {
				dotClr = color.RGBA{255, 50, 50, 255}
			} else {
				dotClr = color.RGBA{255, 200, 200, 255}
			}
		} else {
			alpha := uint8(180 + 40*(g.frame/30%2))
			dotClr = color.RGBA{255, 255, 255, alpha}
		}
		drawDirDot(screen, e.X, e.Y, e.Dir, dotClr, camX, camY)
	}

	// プレイヤー描画
	g.drawActor(screen, &(g.Player.Actor), playerImage, color.White, camX, camY)
	drawDirDot(screen, g.Player.X, g.Player.Y, g.Player.Dir, color.RGBA{255, 255, 255, 255}, camX, camY)

	// 遠隔攻撃
	for _, p := range g.projectiles {
		t := float64(p.Frame) / float64(p.TotalFrames)
		curX := p.StartX + (p.EndX-p.StartX)*t
		curY := p.StartY + (p.EndY-p.StartY)*t
		pRect := ebiten.NewImage(8, 8)
		pRect.Fill(p.Color)
		pOp := &ebiten.DrawImageOptions{}
		pOp.GeoM.Translate(curX-camX-4, curY-camY-4)
		screen.DrawImage(pRect, pOp)
	}
}

func (g *GameScene) drawActor(screen *ebiten.Image, a *Actor, img *ebiten.Image, fallbackColor color.Color, camX, camY float64) {
	if a.DamageAnim > 0 && (a.DamageAnim/4)%2 == 0 {
		return
	}
	ox, oy, _ := charAnim(g.frame, a.AttackAnim, a.Dir)
	if img != nil {
		dirOffset := 0
		switch a.Dir {
		case DirDown:
			dirOffset = 0
		case DirUp:
			dirOffset = 1
		case DirRight:
			dirOffset = 2
		case DirLeft:
			dirOffset = 3
		}
		frame := (g.frame / 10) % 3
		sx, sy := dirOffset*32, frame*32
		rect := image.Rect(sx, sy, sx+32, sy+32)
		sub := img.SubImage(rect).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(a.X*tileSize+ox*2)-camX, float64(a.Y*tileSize+oy*2)-camY)
		screen.DrawImage(sub, op)
	} else {
		rect := ebiten.NewImage(tileSize, tileSize)
		rect.Fill(fallbackColor)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(a.X*tileSize+ox*2)-camX, float64(a.Y*tileSize+oy*2)-camY)
		screen.DrawImage(rect, op)
	}
}

func (g *GameScene) drawUI(screen *ebiten.Image, offsetY float64) {
	screen.Fill(color.RGBA{0, 0, 0, 255})
	statusY := offsetY + 8
	hpClr := color.RGBA{200, 200, 200, 255}
	if g.Player.HP <= (g.Player.MaxHP)/4 {
		hpClr = color.RGBA{255, 80, 80, 255}
	}
	wt := g.currentWeight()
	drawText(screen, fmt.Sprintf("フロア: %d  ターン: %d", g.floor+1, g.turnCount), fontFace12, 8, int(statusY), color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, fmt.Sprintf("Lv: %d  XP: %d / %d", g.Player.Level, g.Player.XP, g.Player.XPToNext), fontFace12, 160, int(statusY), color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, fmt.Sprintf("HP: %d / %d  MP: %d / %d", g.Player.HP, g.Player.MaxHP, g.Player.MP, g.Player.MaxMP), fontFace12, 8, int(statusY+16), hpClr, false)
	drawText(screen, fmt.Sprintf("ATK: %d  DEF: %d 重量: %d / %d", 1+g.Player.Str+g.equippedPhyAtk(), g.equippedPhyDef()+g.Player.Vit, wt, maxCarryWeight), fontFace12, 160, int(statusY+16), color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, fmt.Sprintf("Str:%d Wis:%d Fai:%d Vit:%d Agi:%d Luk:%d", g.Player.Str, g.Player.Wis, g.Player.Fai, g.Player.Vit, g.Player.Agi, g.Player.Luk), fontFace12, 8, int(statusY+32), color.RGBA{150, 150, 150, 255}, false)
	drawText(screen, g.message, fontFace14, 8, int(statusY+48), color.RGBA{255, 255, 255, 255}, false)
	drawText(screen, "H: ヘルプ", fontFace12, screenWidth-80, int(statusY), color.RGBA{100, 100, 100, 255}, false)
}

func (g *GameScene) drawOverlays(screen *ebiten.Image) {
	if g.menuOpen {
		const (
			rowH      = 32
			padX      = 48
			padRight  = 16
			headerH   = 44
			footerH   = 28
			minPanelW = 240
			maxPanelW = screenWidth - 20
		)
		panelW := minPanelW
		for _, item := range g.menuItems {
			w := measureText(item.label, fontFace14) + padX + padRight
			if w > panelW {
				panelW = w
			}
		}
		panelH := headerH + len(g.menuItems)*rowH + footerH
		if panelH > screenHeight-40 {
			panelH = screenHeight - 40
		}
		visibleRows := (panelH - headerH - footerH) / rowH
		if visibleRows < 1 {
			visibleRows = 1
		}
		scrollOffset := g.menuCursor - visibleRows + 1
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		if g.menuCursor < scrollOffset {
			scrollOffset = g.menuCursor
		}
		panelX, panelY := (screenWidth-panelW)/2, (screenHeight-panelH)/2
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
			}
			drawText(screen, item.label, fontFace14, panelX+30, y, clr, false)
		}
		drawText(screen, "W/S: 選択  Enter: 決定  X/Esc: 閉じる", fontFace12, screenWidth/2, panelY+panelH-20, color.RGBA{120, 120, 120, 255}, true)
	}
	if g.confirmQuit {
		const (
			dW = 300
			dH = 80
			dX = (screenWidth - dW) / 2
			dY = (screenHeight - dH) / 2
		)
		drawPanel(screen, dX, dY, dW, dH, color.RGBA{30, 20, 20, 255}, color.RGBA{200, 100, 100, 255})
		drawText(screen, "タイトルへ戻りますか？", fontFace14, screenWidth/2, dY+14, color.RGBA{255, 220, 220, 255}, true)
		drawText(screen, "Enter/Y: はい    Esc/N: いいえ", fontFace12, screenWidth/2, dY+46, color.RGBA{180, 180, 180, 255}, true)
	}
	if g.playState == StateWin || g.playState == StateDead {
		const (
			dW = 360
			dH = 120
			dX = (screenWidth - dW) / 2
			dY = (screenHeight - dH) / 2
		)
		var pClr, bClr color.RGBA
		var title, sub string
		if g.playState == StateWin {
			pClr = color.RGBA{20, 40, 20, 255}
			bClr = color.RGBA{100, 255, 100, 255}
			title = "--- BEAT THE GAME ---"
			sub = fmt.Sprintf("スコア: %d", g.calcScore())
		} else {
			pClr = color.RGBA{40, 20, 20, 255}
			bClr = color.RGBA{255, 100, 100, 255}
			title = "--- GAME OVER ---"
			sub = "あなたは力尽きた..."
		}
		drawPanel(screen, dX, dY, dW, dH, pClr, bClr)
		drawText(screen, title, fontFace14, screenWidth/2, dY+20, bClr, true)
		drawText(screen, sub, fontFace12, screenWidth/2, dY+55, color.RGBA{220, 220, 220, 255}, true)
		drawText(screen, "R: 再挑戦    Esc: タイトルへ", fontFace12, screenWidth/2, dY+dH-25, color.RGBA{180, 180, 180, 255}, true)
	}

	// デバッグ情報（ビルドタグ debug が有効なときのみ表示）
	g.drawDebug(screen)
}
