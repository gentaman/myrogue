package game

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	creditLineHeight  = 28
	creditScrollSpeed = 1.0 // ピクセル/フレーム
)

// クレジット画面
type CreditScene struct {
	prev    Scene
	lines   []string
	scrollY float64 // 上方向への累計スクロール量
}

func NewCreditScene(prev Scene) *CreditScene {
	lines := strings.Split(strings.ReplaceAll(creditsText, "\r\n", "\n"), "\n")
	return &CreditScene{prev: prev, lines: lines, scrollY: 0}
}

func (s *CreditScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyC) {
		return s.prev, nil
	}
	s.scrollY += creditScrollSpeed
	return nil, nil
}

func (s *CreditScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{5, 5, 20, 255})

	totalH := float64(len(s.lines) * creditLineHeight)
	// 全テキストが流れ終わったらループ
	if s.scrollY > totalH+float64(screenHeight) {
		s.scrollY = 0
	}

	startY := float64(screenHeight) - s.scrollY
	for i, line := range s.lines {
		y := startY + float64(i*creditLineHeight)
		if y < -float64(creditLineHeight) || y > float64(screenHeight) {
			continue
		}
		clr := color.RGBA{200, 200, 200, 255}
		face := fontFace14
		if line == "---" {
			// 区切り線
			sep := ebiten.NewImage(400, 1)
			sep.Fill(color.RGBA{80, 80, 120, 255})
			sepOp := &ebiten.DrawImageOptions{}
			sepOp.GeoM.Translate(float64(screenWidth/2-200), y+float64(creditLineHeight)/2)
			screen.DrawImage(sep, sepOp)
			continue
		}
		// 先頭行（タイトル）と見出し行を強調
		if i == 0 {
			face = fontFace20
			clr = color.RGBA{255, 220, 100, 255}
		} else if len(line) > 0 && line[0] != ' ' && line != "" {
			// 大文字始まりのセクション見出し
			allUpper := true
			for _, r := range line {
				if r >= 'a' && r <= 'z' {
					allUpper = false
					break
				}
			}
			if allUpper && len(line) > 2 {
				face = fontFace14
				clr = color.RGBA{150, 220, 255, 255}
			}
		}
		drawText(screen, line, face, screenWidth/2, int(y), clr, true)
	}

	drawText(screen, "Esc / C : 戻る", fontFace12, screenWidth/2, screenHeight-20, color.RGBA{100, 100, 100, 255}, true)
}
