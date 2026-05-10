package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type ChestScene struct {
	game        *GameScene
	chestIdx    int
	focusLeft   bool // true: Chest, false: Player
	leftCursor  int
	rightCursor int
	leftOffset  int
	rightOffset int
}

func (s *ChestScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return s.game, nil
	}

	chest := &s.game.mapItems[s.chestIdx]
	player := &s.game.Player

	// 左右のフォーカス切り替え
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		s.focusLeft = !s.focusLeft
	}

	// カーソル移動
	if s.focusLeft {
		n := len(chest.Inventory)
		if n > 0 {
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
				s.leftCursor = (s.leftCursor - 1 + n) % n
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
				s.leftCursor = (s.leftCursor + 1) % n
			}
		}
	} else {
		n := len(player.Inventory)
		if n > 0 {
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
				s.rightCursor = (s.rightCursor - 1 + n) % n
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
				s.rightCursor = (s.rightCursor + 1) % n
			}
		}
	}

	// アイテム移動 (Enter)
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if s.focusLeft {
			// Chest -> Player
			if len(chest.Inventory) > 0 {
				entry := chest.Inventory[s.leftCursor]
				// プレイヤーの重量チェック
				if s.game.addInventory(player, entry.kind, entry.Durability, entry.obtainedSeed, entry.obtainedFloor) {
					// 成功したらチェストから消す
					chest.Inventory = append(chest.Inventory[:s.leftCursor], chest.Inventory[s.leftCursor+1:]...)
					if s.leftCursor >= len(chest.Inventory) && s.leftCursor > 0 {
						s.leftCursor--
					}
				} else {
					s.game.pushMessage("重すぎて持てない！")
				}
			}
		} else {
			// Player -> Chest
			if len(player.Inventory) > 0 {
				entry := player.Inventory[s.rightCursor]
				if entry.Equipped {
					s.game.pushMessage("装備中は入れられない。")
				} else {
					// チェストへ移動（重量制限なし）
					chest.Inventory = append(chest.Inventory, entry)
					player.Inventory = append(player.Inventory[:s.rightCursor], player.Inventory[s.rightCursor+1:]...)
					if s.rightCursor >= len(player.Inventory) && s.rightCursor > 0 {
						s.rightCursor--
					}
				}
			}
		}
	}

	return nil, nil
}

func (s *ChestScene) Draw(screen *ebiten.Image) {
	s.game.Draw(screen)

	// 半透明オーバーレイ
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	const (
		panelW  = 280
		panelH  = 340
		gap     = 20
		panelY  = (screenHeight - panelH) / 2
		leftX   = (screenWidth - panelW*2 - gap) / 2
		rightX  = leftX + panelW + gap
		rowH    = 24
		headerH = 40
	)

	chest := &s.game.mapItems[s.chestIdx]
	player := &s.game.Player

	// チェストパネル
	lClr := color.RGBA{60, 60, 100, 255}
	if s.focusLeft {
		lClr = color.RGBA{100, 100, 200, 255}
	}
	drawPanel(screen, leftX, panelY, panelW, panelH, color.RGBA{20, 20, 30, 255}, lClr)
	drawText(screen, "宝箱の中身", fontFace14, leftX+panelW/2, panelY+12, color.RGBA{255, 220, 100, 255}, true)

	for i, entry := range chest.Inventory {
		y := panelY + headerH + i*rowH
		clr := color.RGBA{180, 180, 180, 255}
		if s.focusLeft && i == s.leftCursor {
			drawText(screen, "▶", fontFace12, leftX+10, y, color.RGBA{255, 255, 255, 255}, false)
			clr = color.RGBA{255, 255, 255, 255}
		}
		drawText(screen, itemDefs[entry.kind].entryName(entry), fontFace12, leftX+30, y, clr, false)
	}

	// プレイヤーパネル
	rClr := color.RGBA{60, 60, 100, 255}
	if !s.focusLeft {
		rClr = color.RGBA{100, 100, 200, 255}
	}
	drawPanel(screen, rightX, panelY, panelW, panelH, color.RGBA{20, 20, 30, 255}, rClr)
	drawText(screen, fmt.Sprintf("あなたの持物 (%d/%d)", s.game.currentWeight(), maxCarryWeight), fontFace14, rightX+panelW/2, panelY+12, color.RGBA{150, 255, 150, 255}, true)

	for i, entry := range player.Inventory {
		y := panelY + headerH + i*rowH
		clr := color.RGBA{180, 180, 180, 255}
		if !s.focusLeft && i == s.rightCursor {
			drawText(screen, "▶", fontFace12, rightX+10, y, color.RGBA{255, 255, 255, 255}, false)
			clr = color.RGBA{255, 255, 255, 255}
		}
		drawText(screen, itemDefs[entry.kind].entryName(entry), fontFace12, rightX+30, y, clr, false)
	}

	drawText(screen, "Left/Right/Tab: 左右切替  Enter: アイテム移動  Space: 閉じる", fontFace12, screenWidth/2, panelY+panelH+15, color.RGBA{255, 255, 255, 255}, true)
}
