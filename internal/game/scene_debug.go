//go:build debug

package game

import (
	"fmt"
	"image/color"
	"reflect"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type DebugScene struct {
	game         *GameScene
	scrollOffset int
	lines        []string
	category     int // 0: Player, 1: Enemies, 2: Map, 3: Global
}

func (s *DebugScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		return s.game, nil
	}

	// カテゴリ切り替え
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		s.category = (s.category + 1) % 4
		s.refresh()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		s.category = (s.category - 1 + 4) % 4
		s.refresh()
	}

	// スクロール
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 {
		s.scrollOffset -= 3
	} else if wheelY < 0 {
		s.scrollOffset += 3
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		s.scrollOffset--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		s.scrollOffset++
	}
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
	if len(s.lines) > 0 && s.scrollOffset >= len(s.lines) {
		s.scrollOffset = len(s.lines) - 1
	}

	return nil, nil
}

func (s *DebugScene) refresh() {
	s.lines = []string{}
	s.scrollOffset = 0

	switch s.category {
	case 0: // Player
		s.lines = append(s.lines, "=== PLAYER DETAILS ===")
		s.inspectStruct(s.game.Player)
	case 1: // Enemies
		s.lines = append(s.lines, fmt.Sprintf("=== ENEMIES (%d) ===", len(s.game.enemies)))
		for i := range s.game.enemies {
			e := &s.game.enemies[i]
			s.lines = append(s.lines, fmt.Sprintf("[%d] %s (ID:%d)", i, enemyDefs[e.kind].name, e.ID))
			s.inspectStruct(*e)
			s.lines = append(s.lines, "--------------------")
		}
	case 2: // Map & World
		s.lines = append(s.lines, "=== MAP & WORLD ===")
		s.lines = append(s.lines, fmt.Sprintf("Floor: %d", s.game.floor))
		s.lines = append(s.lines, fmt.Sprintf("Seed: %d", s.game.mapSeed))
		s.lines = append(s.lines, fmt.Sprintf("TurnCount: %d", s.game.turnCount))
		s.lines = append(s.lines, fmt.Sprintf("Rooms: %d", len(s.game.rooms)))
		s.lines = append(s.lines, fmt.Sprintf("MapItems: %d", len(s.game.mapItems)))
		for i, mi := range s.game.mapItems {
			s.lines = append(s.lines, fmt.Sprintf("  Item %d at (%d, %d): %d items", i, mi.X, mi.Y, len(mi.Inventory)))
		}
	case 3: // Global / Systems
		s.lines = append(s.lines, "=== GLOBAL STATS ===")
		s.lines = append(s.lines, fmt.Sprintf("Frame: %d", s.game.frame))
		s.lines = append(s.lines, fmt.Sprintf("AnimSpeed: %.2f", s.game.AnimSpeed))
		s.lines = append(s.lines, fmt.Sprintf("PlayState: %d", s.game.playState))
		s.lines = append(s.lines, fmt.Sprintf("TurnState: %d", s.game.turnState))
	}
}

func (s *DebugScene) inspectStruct(v interface{}) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fVal := val.Field(i)

		// スライスやマップは簡略化
		strVal := ""
		switch fVal.Kind() {
		case reflect.Slice:
			strVal = fmt.Sprintf("(Slice len:%d)", fVal.Len())
		case reflect.Map:
			strVal = fmt.Sprintf("(Map len:%d)", fVal.Len())
		case reflect.Struct:
			strVal = "{...}"
		default:
			if fVal.CanInterface() {
				strVal = fmt.Sprintf("%v", fVal.Interface())
			} else {
				strVal = "<unexported>"
			}
		}

		s.lines = append(s.lines, fmt.Sprintf("  %s: %s", field.Name, strVal))
	}
}

func (s *DebugScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 20, 30, 255})

	// Header
	cats := []string{"[Player]", "[Enemies]", "[Map]", "[Global]"}
	headerX := 10
	for i, c := range cats {
		clr := color.RGBA{150, 150, 150, 255}
		if i == s.category {
			clr = color.RGBA{255, 255, 255, 255}
		}
		drawText(screen, c, fontFace12, headerX, 20, clr, false)
		headerX += 80
	}
	drawText(screen, "TAB/Arrows to Switch", fontFace12, screenWidth-70, 20, color.RGBA{100, 255, 100, 255}, true)
	drawText(screen, " Esc/F1 to Exit", fontFace12, screenWidth-50, 30, color.RGBA{100, 255, 100, 255}, true)

	// Content
	lineH := 16
	startY := 45
	visibleLines := (screenHeight - startY) / lineH

	for i := 0; i < visibleLines; i++ {
		idx := i + s.scrollOffset
		if idx >= len(s.lines) {
			break
		}
		line := s.lines[idx]

		// インデントなどに応じた色分け
		var clr color.Color = color.White
		if strings.HasPrefix(line, "===") {
			clr = color.RGBA{255, 200, 0, 255}
		} else if strings.HasPrefix(line, "  ") {
			clr = color.RGBA{200, 200, 200, 255}
		}

		drawText(screen, line, fontFace12, 10, startY+i*lineH, clr, false)
	}

	// Scrollbar
	if len(s.lines) > visibleLines {
		barH := float64(visibleLines) / float64(len(s.lines)) * float64(screenHeight-startY)
		barY := float64(s.scrollOffset) / float64(len(s.lines)) * float64(screenHeight-startY)
		ebiten.NewImage(4, int(barH))
		barImg := ebiten.NewImage(4, int(barH))
		barImg.Fill(color.RGBA{100, 100, 100, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(screenWidth-6), float64(startY)+barY)
		screen.DrawImage(barImg, op)
	}
}
