package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type invSubMenu int

const (
	invSubNone invSubMenu = iota
	invSubUse
	invSubDrop
	invSubDesc
)

// InventoryScene はインベントリ画面を表す
type InventoryScene struct {
	game         *GameScene
	cursor       int
	subMenu      invSubMenu
	subCursor    int
	scrollOffset int
}

var subMenuLabels = []string{"使う", "捨てる", "説明を見る", "キャンセル"}

func (s *InventoryScene) clampCursor() {
	n := len(s.game.inventory)
	if n == 0 {
		s.cursor = 0
		return
	}
	if s.cursor >= n {
		s.cursor = n - 1
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
}

func (s *InventoryScene) syncScroll(visibleRows int) {
	if s.cursor < s.scrollOffset {
		s.scrollOffset = s.cursor
	}
	if s.cursor >= s.scrollOffset+visibleRows {
		s.scrollOffset = s.cursor - visibleRows + 1
	}
}

func (s *InventoryScene) calcLayout() (panelW, panelH, visibleRows int) {
	const (
		rowH      = 32
		padX      = 58
		padRight  = 16
		headerH   = 44
		footerH   = 28
		minW      = 300
		maxW      = screenWidth - 20
	)
	areaH := mapHeight * tileSize
	panelW = minW
	hintW := measureText("W/S: 選択  Enter: アクション  I/Esc: 閉じる", fontFace12)
	if hintW+padRight > panelW {
		panelW = hintW + padRight
	}
	for _, entry := range s.game.inventory {
		def := &itemDefs[entry.kind]
		label := fmt.Sprintf("%s  x%d  (重%d)", def.entryName(entry), entry.count, def.weight)
		w := measureText(label, fontFace14) + padX + padRight
		if w > panelW {
			panelW = w
		}
	}
	if panelW > maxW {
		panelW = maxW
	}
	n := len(s.game.inventory)
	if n == 0 {
		n = 1
	}
	needH := headerH + n*rowH + footerH
	panelH = needH
	if panelH > areaH-8 {
		panelH = areaH - 8
	}
	visibleRows = (panelH - headerH - footerH) / rowH
	if visibleRows < 1 {
		visibleRows = 1
	}
	return
}

func (s *InventoryScene) Update() (Scene, error) {
	inv := s.game.inventory

	// 説明ウィンドウ表示中
	if s.subMenu == invSubDesc {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.subMenu = invSubNone
		}
		return nil, nil
	}

	// サブメニュー表示中
	if s.subMenu != invSubNone {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			s.subMenu = invSubNone
			return nil, nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			s.subCursor = (s.subCursor - 1 + len(subMenuLabels)) % len(subMenuLabels)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			s.subCursor = (s.subCursor + 1) % len(subMenuLabels)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			switch s.subCursor {
			case 0: // 使う
				if s.game.useInventoryItem(s.cursor) {
					s.clampCursor()
					s.subMenu = invSubNone
					s.game.turnCount++
					s.game.moveEnemies()
					if s.game.playState == StateDead {
						return s.game, nil
					}
					s.game.trySpawnEnemyPerTurn()
					return s.game, nil
				}
			case 1: // 捨てる
				s.game.dropInventoryItem(s.cursor)
				s.clampCursor()
				s.subMenu = invSubNone
			case 2: // 説明を見る
				s.subMenu = invSubDesc
			case 3: // キャンセル
				s.subMenu = invSubNone
			}
		}
		return nil, nil
	}


	// メインリスト操作
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyI) {
		s.game.menuOpen = false
		return s.game, nil
	}
	if len(inv) == 0 {
		return nil, nil
	}
	_, _, visibleRows := s.calcLayout()
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.cursor = (s.cursor - 1 + len(inv)) % len(inv)
		s.syncScroll(visibleRows)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.cursor = (s.cursor + 1) % len(inv)
		s.syncScroll(visibleRows)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.subMenu = invSubUse
		s.subCursor = 0
	}
	return nil, nil
}

func (s *InventoryScene) Draw(screen *ebiten.Image) {
	s.game.Draw(screen)

	const (
		rowH    = 32
		headerH = 44
		footerH = 28
	)

	panelW, panelH, visibleRows := s.calcLayout()
	panelX := (screenWidth - panelW) / 2
	panelY := (mapHeight*tileSize - panelH) / 2

	drawPanel(screen, panelX, panelY, panelW, panelH, color.RGBA{15, 20, 40, 255}, color.RGBA{80, 180, 100, 255})

	drawText(screen, "アイテム", fontFace14, screenWidth/2, panelY+12, color.RGBA{150, 255, 150, 255}, true)

	inv := s.game.inventory
	if len(inv) == 0 {
		drawText(screen, "アイテムを持っていない", fontFace14, screenWidth/2, panelY+panelH/2-10, color.RGBA{120, 120, 120, 255}, true)
	} else {
		// スクロールインジケーター
		if s.scrollOffset > 0 {
			drawText(screen, "▲", fontFace12, panelX+panelW/2, panelY+headerH-14, color.RGBA{150, 200, 150, 255}, true)
		}
		if s.scrollOffset+visibleRows < len(inv) {
			drawText(screen, "▼", fontFace12, panelX+panelW/2, panelY+headerH+visibleRows*rowH, color.RGBA{150, 200, 150, 255}, true)
		}

		for i, entry := range inv {
			row := i - s.scrollOffset
			if row < 0 || row >= visibleRows {
				continue
			}
			y := panelY + headerH + row*rowH
			def := itemDefs[entry.kind]
			clr := color.RGBA{200, 200, 200, 255}
			if i == s.cursor {
				hl := ebiten.NewImage(panelW-8, 26)
				hl.Fill(color.RGBA{40, 80, 50, 255})
				hlOp := &ebiten.DrawImageOptions{}
				hlOp.GeoM.Translate(float64(panelX+4), float64(y-4))
				screen.DrawImage(hl, hlOp)
				clr = color.RGBA{255, 255, 255, 255}
				drawText(screen, "▶", fontFace12, panelX+14, y, clr, false)
			}
			dot := ebiten.NewImage(8, 8)
			dot.Fill(def.clr)
			dotOp := &ebiten.DrawImageOptions{}
			dotOp.GeoM.Translate(float64(panelX+30), float64(y+3))
			screen.DrawImage(dot, dotOp)

			label := fmt.Sprintf("%s  x%d  (重%d)", def.entryName(entry), entry.count, def.weight)
			labelMaxW := panelX + panelW - 16 - (panelX + 44)
			label = truncateText(label, fontFace14, labelMaxW)
			drawText(screen, label, fontFace14, panelX+44, y, clr, false)
		}
	}

	if s.subMenu == invSubDesc {
		s.drawDescWindow(screen, panelX, panelY, panelW, panelH)
	} else if s.subMenu != invSubNone {
		s.drawSubMenu(screen, panelX, panelY, panelW)
	}

	hint := "W/S: 選択  Enter: アクション  I/Esc: 閉じる"
	if s.subMenu == invSubDesc {
		hint = "Enter/Esc: 閉じる"
	} else if s.subMenu != invSubNone {
		hint = "W/S: 選択  Enter: 決定  Esc: キャンセル"
	}
	drawText(screen, hint, fontFace12, screenWidth/2, panelY+panelH-20, color.RGBA{100, 120, 100, 255}, true)
}

func (s *InventoryScene) drawSubMenu(screen *ebiten.Image, panelX, panelY, panelW int) {
	const (
		subRowH   = 26
		subPadX   = 36
		subPadR   = 16
		subPadTop = 14
		subPadBot = 10
		minSubW   = 120
	)
	subW := minSubW
	for _, label := range subMenuLabels {
		w := measureText(label, fontFace14) + subPadX + subPadR
		if w > subW {
			subW = w
		}
	}
	if subW > screenWidth-20 {
		subW = screenWidth - 20
	}
	subH := subPadTop + len(subMenuLabels)*subRowH + subPadBot

	subX := panelX + (panelW-subW)/2
	row := s.cursor - s.scrollOffset
	if row < 0 {
		row = 0
	}
	subY := panelY + 44 + row*32 - 8
	if subY+subH > mapHeight*tileSize {
		subY = mapHeight*tileSize - subH - 4
	}
	if subY < 0 {
		subY = 4
	}

	drawPanel(screen, subX, subY, subW, subH, color.RGBA{30, 30, 60, 255}, color.RGBA{100, 150, 255, 255})

	for i, label := range subMenuLabels {
		y := subY + subPadTop + i*subRowH
		clr := color.RGBA{200, 200, 200, 255}
		if i == s.subCursor {
			hl := ebiten.NewImage(subW-8, 22)
			hl.Fill(color.RGBA{60, 60, 120, 255})
			hlOp := &ebiten.DrawImageOptions{}
			hlOp.GeoM.Translate(float64(subX+4), float64(y-3))
			screen.DrawImage(hl, hlOp)
			clr = color.RGBA{255, 255, 255, 255}
			drawText(screen, "▶", fontFace12, subX+10, y, clr, false)
		}
		subLabelMaxW := subW - 26 - 8
		drawText(screen, truncateText(label, fontFace14, subLabelMaxW), fontFace14, subX+26, y, clr, false)
	}
}

func (s *InventoryScene) drawDescWindow(screen *ebiten.Image, panelX, panelY, panelW, panelH int) {
	if s.cursor < 0 || s.cursor >= len(s.game.inventory) {
		return
	}
	entry := s.game.inventory[s.cursor]
	def := itemDefs[entry.kind]

	const (
		dPad  = 16
		lineH = 20
	)
	dW := panelW - dPad*2
	if dW < 200 {
		dW = 200
	}
	maxLineW := dW - dPad*2

	// 説明文を maxLineW に収まるよう行に分割
	lines := wrapText(def.desc, fontFace12, maxLineW)

	titleH := 28
	dH := titleH + len(lines)*lineH + dPad*2
	dX := panelX + dPad
	dY := panelY + (panelH-dH)/2
	if dY < panelY+4 {
		dY = panelY + 4
	}

	drawPanel(screen, dX, dY, dW, dH, color.RGBA{20, 20, 50, 255}, color.RGBA{150, 200, 255, 255})
	drawText(screen, def.entryName(entry), fontFace14, dX+dW/2, dY+8, color.RGBA{200, 230, 255, 255}, true)

	// ステータス行
	statsY := dY + titleH
	drawText(screen, fmt.Sprintf("効果: %s  重量: %d", def.effectDesc, def.weight), fontFace12, dX+dPad, statsY, color.RGBA{160, 220, 160, 255}, false)

	for i, line := range lines {
		drawText(screen, line, fontFace12, dX+dPad, statsY+lineH+i*lineH, color.RGBA{200, 200, 200, 255}, false)
	}
}

// wrapText は str を face で描画したとき maxWidth px を超えないよう行分割する
func wrapText(str string, face *text.GoTextFace, maxWidth int) []string {
	var lines []string
	runes := []rune(str)
	start := 0
	for start < len(runes) {
		end := len(runes)
		// バイナリサーチで収まる最大位置を探す
		lo, hi := start+1, len(runes)
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if measureText(string(runes[start:mid]), face) <= maxWidth {
				lo = mid
			} else {
				hi = mid - 1
			}
		}
		end = lo
		lines = append(lines, string(runes[start:end]))
		start = end
	}
	return lines
}

// drawPanel は角丸なしのパネルと枠線を描画するヘルパー
func drawPanel(screen *ebiten.Image, x, y, w, h int, bg, border color.RGBA) {
	panel := ebiten.NewImage(w, h)
	panel.Fill(bg)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(panel, op)

	for _, r := range []struct{ x, y, w, h int }{
		{x, y, w, 2},
		{x, y + h - 2, w, 2},
		{x, y, 2, h},
		{x + w - 2, y, 2, h},
	} {
		b := ebiten.NewImage(r.w, r.h)
		b.Fill(border)
		bOp := &ebiten.DrawImageOptions{}
		bOp.GeoM.Translate(float64(r.x), float64(r.y))
		screen.DrawImage(b, bOp)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
