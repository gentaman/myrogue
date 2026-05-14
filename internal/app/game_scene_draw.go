package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

func (g *GameScene) cameraOffset() (float64, float64) {
	pPos := g.playerPos()
	cx := float64(pPos.X*TileSize+TileSize/2) - float64(ScreenWidth)/2
	cy := float64(pPos.Y*TileSize+TileSize/2) - float64(GameplayAreaHeight)/2
	maxX := float64(world.MapWidth*TileSize - ScreenWidth)
	maxY := float64(world.MapHeight*TileSize - GameplayAreaHeight)
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	if cx > maxX {
		cx = maxX
	}
	if cy > maxY {
		cy = maxY
	}
	return cx, cy
}

func (g *GameScene) charAnimOffset(attackAnim int, dir component.Dir) (int, int) {
	if attackAnim > 0 {
		shift := 1
		if attackAnim > component.AttackAnimFrames/2 {
			shift = 3
		}
		dx, dy := dir.Delta()
		return dx * shift, dy * shift
	}
	return 0, 0
}

func (g *GameScene) Draw(r Renderer) {
	r.Clear(20, 20, 20, 255)
	g.drawWorld(r)
	g.drawUI(r)
	g.drawOverlays(r)
	g.drawDebugHUD(r)
}

func (g *GameScene) drawWorld(r Renderer) {
	camX, camY := g.cameraOffset()

	for x := 0; x < world.MapWidth; x++ {
		for y := 0; y < world.MapHeight; y++ {
			if !g.World.Explored[x][y] && !g.Debug.RevealMap {
				continue
			}
			vis := g.World.Visible[x][y] || g.Debug.RevealMap
			var clr Color
			if vis {
				switch g.World.Tiles[x][y] {
				case world.Wall:
					clr = Color{100, 100, 110, 255}
				case world.Floor:
					clr = Color{60, 60, 65, 255}
				case world.Stairs:
					clr = Color{0, 255, 255, 255}
				case world.StairsDown:
					clr = Color{255, 140, 0, 255}
				case world.StairsUp:
					clr = Color{0, 220, 180, 255}
				}
			} else {
				switch g.World.Tiles[x][y] {
				case world.Wall:
					clr = Color{50, 50, 55, 255}
				case world.Floor:
					clr = Color{28, 28, 30, 255}
				case world.Stairs:
					clr = Color{0, 120, 120, 255}
				case world.StairsDown:
					clr = Color{140, 70, 0, 255}
				case world.StairsUp:
					clr = Color{0, 110, 90, 255}
				}
			}
			sx := float64(x*TileSize+1) - camX
			sy := float64(y*TileSize+1) - camY
			if sx > float64(ScreenWidth) || sy > float64(GameplayAreaHeight) || sx < -float64(TileSize) || sy < -float64(TileSize) {
				continue
			}
			r.DrawRect(int(sx), int(sy), TileSize-2, TileSize-2, clr)
			if g.Debug.ShowGrid {
				r.DrawRect(int(sx), int(sy), TileSize-2, 1, Color{40, 40, 60, 100})
				r.DrawRect(int(sx), int(sy), 1, TileSize-2, Color{40, 40, 60, 100})
			}
		}
	}

	for _, it := range g.World.Items {
		if !g.World.Explored[it.X][it.Y] && !g.Debug.RevealMap {
			continue
		}
		var clr Color
		if len(it.Inventory) > 1 {
			clr = Color{180, 130, 40, 255}
		} else if len(it.Inventory) == 1 {
			clr = Color{200, 200, 100, 255}
		} else {
			continue
		}
		if !g.World.Visible[it.X][it.Y] {
			clr = Color{clr[0] / 2, clr[1] / 2, clr[2] / 2, 255}
		}
		ix := float64(it.X*TileSize+8) - camX
		iy := float64(it.Y*TileSize+8) - camY
		r.DrawRect(int(ix), int(iy), TileSize-16, TileSize-16, clr)
	}

	// Draw entities
	g.Factions.Each(func(id entity.ID, fc *component.FactionComp) {
		if !g.Entities.IsAlive(id) {
			return
		}
		pos := g.GetPosition(id)
		if pos == nil {
			return
		}
		if !g.World.Visible[pos.X][pos.Y] && !g.Debug.RevealMap && id != g.Player {
			return
		}

		anim := g.GetAnim(id)
		if anim != nil && anim.DamageAnim > 0 && (int(anim.DamageAnim)/4)%2 == 0 {
			// Damage flash
			clr := hexToColor(ColorNone) // Default
			if anim.DamageColor != "" {
				clr = hexToColor(anim.DamageColor)
			}
			sx := float64(pos.X*TileSize) - camX
			sy := float64(pos.Y*TileSize) - camY
			r.DrawRect(int(sx), int(sy), TileSize-1, TileSize-1, clr)
			return
		}

		attackAnim := 0
		if anim != nil {
			attackAnim = int(anim.AttackAnim)
		}
		ox, oy := g.charAnimOffset(attackAnim, pos.Dir)

		// Check for sprite
		appearance, hasApp := g.Appearances.Get(id)
		if hasApp && appearance.HasSprite {
			// Sprite sheet: dir mapping (Down=0, Up=1, Right=2, Left=3)
			dirIdx := 0
			switch pos.Dir {
			case component.DirDown:
				dirIdx = 0
			case component.DirUp:
				dirIdx = 1
			case component.DirRight:
				dirIdx = 2
			case component.DirLeft:
				dirIdx = 3
			}
			frame := (g.Frame / 10) % 3
			sx := float64(pos.X*TileSize+ox*2) - camX
			sy := float64(pos.Y*TileSize+oy*2) - camY
			r.DrawSprite(appearance.DefID, frame, dirIdx, sx, sy)
		} else {
			var clr Color
			switch fc.Faction {
			case component.FactionPlayer:
				clr = Color{255, 255, 255, 255}
			case component.FactionAlly:
				clr = Color{100, 200, 100, 255}
			case component.FactionEnemy:
				clr = Color{200, 80, 80, 255}
			}

			if hasApp && appearance.ColorHex != "" {
				clr = hexToColor(appearance.ColorHex)
			}

			sx := float64(pos.X*TileSize+ox*2) - camX
			sy := float64(pos.Y*TileSize+oy*2) - camY
			r.DrawRect(int(sx), int(sy), TileSize-1, TileSize-1, clr)
		}

		// Direction dot
		var dotClr Color
		switch fc.Faction {
		case component.FactionPlayer:
			dotClr = Color{255, 255, 255, 255}
		case component.FactionAlly:
			alpha := uint8(180 + 40*((g.Frame/30)%2))
			dotClr = Color{100, 255, 100, alpha}
		case component.FactionEnemy:
			aiComp := g.GetAI(id)
			if aiComp != nil && aiComp.State > 0 {
				if (g.Frame/5)%2 == 0 {
					dotClr = Color{255, 50, 50, 255}
				} else {
					dotClr = Color{255, 200, 200, 255}
				}
			} else {
				alpha := uint8(180 + 40*((g.Frame/30)%2))
				dotClr = Color{255, 255, 255, alpha}
			}
		}
		g.drawDirDot(r, pos.X, pos.Y, pos.Dir, dotClr, camX, camY)

		if g.Debug.ShowEntityID {
			ex := float64(pos.X*TileSize) - camX
			ey := float64(pos.Y*TileSize) - camY - 4
			r.DrawText(fmt.Sprintf("%d", id), 9, int(ex), int(ey), Color{255, 255, 0, 200}, false)
		}
	})

	// Projectiles
	for _, p := range g.AnimQueue.Projectiles {
		if p.IsFlash {
			alpha := uint8(200)
			if (p.Frame/4)%2 == 0 {
				alpha = 100
			}
			clr := hexToColor(p.ColorHex)
			clr[3] = alpha
			px := p.EndX - camX
			py := p.EndY - camY
			r.DrawRect(int(px), int(py), TileSize, TileSize, clr)
			continue
		}
		t := float64(p.Frame) / float64(p.TotalFrames)
		curX := p.StartX + (p.EndX-p.StartX)*t
		curY := p.StartY + (p.EndY-p.StartY)*t
		clr := hexToColor(p.ColorHex)
		if p.ColorHex == "#FF4500" { // Fireball flicker
			if (g.Frame/5)%2 == 0 {
				clr = Color{255, 200, 50, 255} // Yellow-ish
			}
		}
		// Larger projectile (12x12) centered at curX, curY
		r.DrawRect(int(curX-camX-6), int(curY-camY-6), 12, 12, clr)
	}
}

func (g *GameScene) drawDirDot(r Renderer, tileX, tileY int, d component.Dir, clr Color, camX, camY float64) {
	const dotSize = 6
	const half = (TileSize - 1) / 2
	ox, oy := 0, 0
	switch d {
	case component.DirUp:
		ox, oy = half-dotSize/2, 4
	case component.DirDown:
		ox, oy = half-dotSize/2, TileSize-1-dotSize-4
	case component.DirLeft:
		ox, oy = 4, half-dotSize/2
	case component.DirRight:
		ox, oy = TileSize-1-dotSize-4, half-dotSize/2
	}
	sx := float64(tileX*TileSize+ox) - camX
	sy := float64(tileY*TileSize+oy) - camY
	r.DrawRect(int(sx), int(sy), dotSize, dotSize, clr)
}

func (g *GameScene) drawUI(r Renderer) {
	r.DrawRect(0, GameplayAreaHeight, ScreenWidth, ScreenHeight-GameplayAreaHeight, Color{0, 0, 0, 255})
	statusY := GameplayAreaHeight + 8

	hpClr := Color{200, 200, 200, 255}
	stats := g.GetStats(g.Player)
	if stats != nil && stats.HP <= stats.MaxHP/4 {
		hpClr = Color{255, 80, 80, 255}
	}

	r.DrawText(fmt.Sprintf("フロア: %d  ターン: %d", g.World.Floor+1, g.Scheduler.TurnCount), 12, 8, statusY, Color{200, 200, 200, 255}, false)
	if stats != nil {
		r.DrawText(fmt.Sprintf("Lv: %d  XP: %d / %d", stats.Level, stats.XP, stats.XPToNext), 12, 160, statusY, Color{200, 200, 200, 255}, false)
		r.DrawText(fmt.Sprintf("HP: %d / %d  MP: %d / %d", stats.HP, stats.MaxHP, stats.MP, stats.MaxMP), 12, 8, statusY+16, hpClr, false)
		atk := 1 + stats.Str + equippedPhyAtk(g.GetInventory(g.Player), g.registry)
		def := equippedPhyDef(g.GetInventory(g.Player), g.registry) + stats.Vit
		r.DrawText(fmt.Sprintf("ATK: %d  DEF: %d", atk, def), 12, 160, statusY+16, Color{200, 200, 200, 255}, false)
		r.DrawText(fmt.Sprintf("Str:%d Wis:%d Fai:%d Vit:%d Agi:%d Luk:%d", stats.Str, stats.Wis, stats.Fai, stats.Vit, stats.Agi, stats.Luk), 12, 8, statusY+32, Color{150, 150, 150, 255}, false)
	}
	if se, ok := g.StatusEffects.Get(g.Player); ok && len(se.Effects) > 0 {
		statusStr := ""
		for i, eff := range se.Effects {
			if i > 0 {
				statusStr += " "
			}
			statusStr += fmt.Sprintf("[%s:%d]", component.StatusNames[eff.Type], eff.Duration)
		}
		r.DrawText(statusStr, 11, 340, statusY+16, Color{255, 200, 100, 255}, false)
	}
	r.DrawText(g.Message, 14, 8, statusY+48, Color{255, 255, 255, 255}, false)
	r.DrawText("H: ヘルプ", 12, ScreenWidth-80, statusY, Color{100, 100, 100, 255}, false)
}

func (g *GameScene) drawOverlays(r Renderer) {
	if g.MenuOpen {
		const (
			rowH      = 32
			padX      = 48
			padRight  = 16
			headerH   = 44
			footerH   = 28
			minPanelW = 240
		)
		panelW := minPanelW
		for _, item := range g.MenuItems {
			w := r.MeasureText(item.Label, 14) + padX + padRight
			if w > panelW {
				panelW = w
			}
		}
		panelH := headerH + len(g.MenuItems)*rowH + footerH
		if panelH > ScreenHeight-40 {
			panelH = ScreenHeight - 40
		}
		visibleRows := (panelH - headerH - footerH) / rowH
		if visibleRows < 1 {
			visibleRows = 1
		}
		scrollOffset := g.MenuCursor - visibleRows + 1
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		if g.MenuCursor < scrollOffset {
			scrollOffset = g.MenuCursor
		}
		panelX, panelY := (ScreenWidth-panelW)/2, (ScreenHeight-panelH)/2
		r.DrawPanel(panelX, panelY, panelW, panelH, Color{20, 20, 50, 255}, Color{100, 150, 255, 255})
		r.DrawText("アクション", 14, ScreenWidth/2, panelY+12, Color{255, 220, 100, 255}, true)
		for i, item := range g.MenuItems {
			row := i - scrollOffset
			if row < 0 || row >= visibleRows {
				continue
			}
			y := panelY + headerH + row*rowH
			clr := Color{180, 180, 180, 255}
			if !item.Enabled {
				clr = Color{80, 80, 80, 255}
			}
			if i == g.MenuCursor {
				r.DrawRect(panelX+4, y-4, panelW-8, 26, Color{60, 60, 120, 255})
				clr = Color{255, 255, 255, 255}
			}
			r.DrawText(item.Label, 14, panelX+30, y, clr, false)
		}
		r.DrawText("W/S: 選択  Enter: 決定  X/Esc: 閉じる", 12, ScreenWidth/2, panelY+panelH-20, Color{120, 120, 120, 255}, true)
	}

	if g.ConfirmStair {
		const (
			dW = 340
			dH = 100
			dX = (ScreenWidth - dW) / 2
			dY = (ScreenHeight - dH) / 2
		)
		r.DrawPanel(dX, dY, dW, dH, Color{30, 30, 20, 255}, Color{200, 180, 100, 255})
		r.DrawText("はぐれる仲間がいます！", 14, ScreenWidth/2, dY+14, Color{255, 220, 100, 255}, true)
		names := ""
		for i, n := range g.StairLeft {
			if i > 0 {
				names += "、"
			}
			names += n
		}
		r.DrawText(names+" は離れすぎている", 12, ScreenWidth/2, dY+40, Color{220, 220, 220, 255}, true)
		r.DrawText("それでも移動しますか？", 12, ScreenWidth/2, dY+58, Color{200, 200, 200, 255}, true)
		r.DrawText("Enter/Y: はい    Esc/N: いいえ", 12, ScreenWidth/2, dY+80, Color{180, 180, 180, 255}, true)
	}

	if g.ConfirmQuit {
		const (
			dW = 300
			dH = 80
			dX = (ScreenWidth - dW) / 2
			dY = (ScreenHeight - dH) / 2
		)
		r.DrawPanel(dX, dY, dW, dH, Color{30, 20, 20, 255}, Color{200, 100, 100, 255})
		r.DrawText("タイトルへ戻りますか？", 14, ScreenWidth/2, dY+14, Color{255, 220, 220, 255}, true)
		r.DrawText("Enter/Y: はい    Esc/N: いいえ", 12, ScreenWidth/2, dY+46, Color{180, 180, 180, 255}, true)
	}

	if g.PlayState == StateWin || g.PlayState == StateDead {
		const (
			dW = 360
			dH = 120
			dX = (ScreenWidth - dW) / 2
			dY = (ScreenHeight - dH) / 2
		)
		var pClr, bClr Color
		var title, sub string
		if g.PlayState == StateWin {
			pClr = Color{20, 40, 20, 255}
			bClr = Color{100, 255, 100, 255}
			title = "--- BEAT THE GAME ---"
			sub = "おめでとうございます！"
		} else {
			pClr = Color{40, 20, 20, 255}
			bClr = Color{255, 100, 100, 255}
			title = "--- GAME OVER ---"
			sub = "あなたは力尽きた..."
		}
		r.DrawPanel(dX, dY, dW, dH, pClr, bClr)
		r.DrawText(title, 14, ScreenWidth/2, dY+20, bClr, true)
		r.DrawText(sub, 12, ScreenWidth/2, dY+55, Color{220, 220, 220, 255}, true)
		r.DrawText("R: 再挑戦    Esc: タイトルへ", 12, ScreenWidth/2, dY+dH-25, Color{180, 180, 180, 255}, true)
	}
}

func hexToColor(hex string) Color {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return Color{255, 255, 255, 255}
	}
	parseHex := func(s string) uint8 {
		v := 0
		for _, c := range s {
			v *= 16
			switch {
			case c >= '0' && c <= '9':
				v += int(c - '0')
			case c >= 'a' && c <= 'f':
				v += int(c-'a') + 10
			case c >= 'A' && c <= 'F':
				v += int(c-'A') + 10
			}
		}
		return uint8(v)
	}
	return Color{parseHex(hex[0:2]), parseHex(hex[2:4]), parseHex(hex[4:6]), 255}
}
