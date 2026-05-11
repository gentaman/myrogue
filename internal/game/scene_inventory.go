package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type invSubMenu int

const (
	invSubNone invSubMenu = iota
	invSubUse
	invSubDrop
	invSubDesc
)

type InventoryScene struct {
	game         *GameScene
	target       *Actor // nil の場合はプレイヤー
	cursor       int
	subMenu      invSubMenu
	subCursor    int
	scrollOffset int
}

var subMenuLabels = []string{"使う", "捨てる", "説明を見る", "キャンセル"}

func (s *InventoryScene) getInventory() []InventoryEntry {
	if s.target != nil {
		return s.target.Inventory
	}
	return s.game.Player.Inventory
}

func (s *InventoryScene) clampCursor() {
	inv := s.getInventory()
	n := len(inv)
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
		rowH     = 32
		padX     = 58
		padRight = 16
		headerH  = 44
		footerH  = 28
		minW     = 300
		maxW     = screenWidth - 20
	)
	panelW = minW
	hintW := measureText("W/S: 選択  Enter: アクション  I/Esc: 閉じる", fontFace12)
	if hintW+padRight > panelW {
		panelW = hintW + padRight
	}
	inv := s.getInventory()
	for _, entry := range inv {
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
	n := len(inv)
	if n == 0 {
		n = 1
	}
	needH := headerH + n*rowH + footerH
	panelH = needH
	if panelH > screenHeight-40 {
		panelH = screenHeight - 40
	}
	visibleRows = (panelH - headerH - footerH) / rowH
	if visibleRows < 1 {
		visibleRows = 1
	}
	return
}

func (s *InventoryScene) Update() (Scene, error) {
	inv := s.getInventory()
	if s.subMenu == invSubDesc {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.subMenu = invSubNone
		}
		return nil, nil
	}
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
				// 仲間の場合は「使う」を無効にするか、特別な処理にするか
				// とりあえずプレイヤーのみ「使う」ができることにする
				if s.target != nil {
					s.game.pushMessage("仲間のアイテムは直接使えない。")
					s.subMenu = invSubNone
					return nil, nil
				}
				if s.game.useInventoryItem(s.cursor) {
					s.clampCursor()
					s.subMenu = invSubNone
					s.game.turnCount++
					s.game.turnState = TurnCompanionAct
					s.game.activeEnemyIdx = 0
					return s.game, nil
				}
			case 1: // 捨てる
				if s.target != nil {
					// 仲間のアイテムを捨てる
					s.target.Inventory = append(s.target.Inventory[:s.cursor], s.target.Inventory[s.cursor+1:]...)
					s.game.pushMessage("仲間のアイテムを捨てさせた。")
				} else {
					s.game.dropInventoryItem(s.cursor)
				}
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
	panelY := (screenHeight - panelH) / 2
	drawPanel(screen, panelX, panelY, panelW, panelH, color.RGBA{15, 20, 40, 255}, color.RGBA{80, 180, 100, 255})

	title := "アイテム"
	if s.target != nil {
		title = s.target.GetName() + "の持ち物"
	}
	drawText(screen, title, fontFace14, screenWidth/2, panelY+12, color.RGBA{150, 255, 150, 255}, true)

	inv := s.getInventory()
	if len(inv) == 0 {
		drawText(screen, "アイテムを持っていない", fontFace14, screenWidth/2, panelY+panelH/2-10, color.RGBA{120, 120, 120, 255}, true)
	} else {
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
			drawText(screen, truncateText(label, fontFace14, panelX+panelW-16-(panelX+44)), fontFace14, panelX+44, y, clr, false)
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
	subH := subPadTop + len(subMenuLabels)*subRowH + subPadBot
	subX := panelX + (panelW-subW)/2
	row := s.cursor - s.scrollOffset
	if row < 0 {
		row = 0
	}
	subY := panelY + 44 + row*32 - 8
	if subY+subH > screenHeight-4 {
		subY = screenHeight - subH - 4
	}
	if subY < 4 {
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
		drawText(screen, truncateText(label, fontFace14, subW-26-8), fontFace14, subX+26, y, clr, false)
	}
}

func (s *InventoryScene) drawDescWindow(screen *ebiten.Image, panelX, panelY, panelW, panelH int) {
	inv := s.getInventory()
	if s.cursor < 0 || s.cursor >= len(inv) {
		return
	}
	entry := inv[s.cursor]
	def := itemDefs[entry.kind]
	const (
		dPad  = 16
		lineH = 20
	)
	dW := panelW - dPad*2
	if dW < 200 {
		dW = 200
	}
	lines := wrapText(def.desc, fontFace12, dW-dPad*2)
	titleH := 28
	dH := titleH + len(lines)*lineH + dPad*2
	dX, dY := panelX+dPad, panelY+(panelH-dH)/2
	if dY < panelY+4 {
		dY = panelY + 4
	}
	drawPanel(screen, dX, dY, dW, dH, color.RGBA{20, 20, 50, 255}, color.RGBA{150, 200, 255, 255})
	drawText(screen, def.entryName(entry), fontFace14, dX+dW/2, dY+8, color.RGBA{200, 230, 255, 255}, true)
	statsY := dY + titleH
	drawText(screen, fmt.Sprintf("効果: %s  重量: %d", def.effectDesc, def.weight), fontFace12, dX+dPad, statsY, color.RGBA{160, 220, 160, 255}, false)
	for i, line := range lines {
		drawText(screen, line, fontFace12, dX+dPad, statsY+lineH+i*lineH, color.RGBA{200, 200, 200, 255}, false)
	}
}
