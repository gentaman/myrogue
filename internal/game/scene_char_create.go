package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type charCreateItem int

const (
	charCreateLevel charCreateItem = iota
	charCreateRace
	charCreateStart
)

type CharacterCreateScene struct {
	level      int
	raceIdx    int
	cursor     charCreateItem
	initialStr int
	initialWis int
	initialFai int
	initialVit int
	initialAgi int
	initialLuk int
}

func NewCharacterCreateScene() *CharacterCreateScene {
	return &CharacterCreateScene{
		level:      1,
		raceIdx:    0,
		cursor:     charCreateLevel,
		initialStr: 1,
		initialWis: 1,
		initialFai: 1,
		initialVit: 1,
		initialAgi: 1,
		initialLuk: 1,
	}
}

func (s *CharacterCreateScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return &TitleScene{}, nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.cursor = (s.cursor - 1 + 3) % 3
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.cursor = (s.cursor + 1) % 3
	}

	switch s.cursor {
	case charCreateLevel:
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			if s.level > 1 {
				s.level--
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			if s.level < 100 {
				s.level++
			}
		}
	case charCreateRace:
		n := len(OrderedRaces)
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			s.raceIdx = (s.raceIdx - 1 + n) % n
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			s.raceIdx = (s.raceIdx + 1) % n
		}
	case charCreateStart:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			return s.startGame(), nil
		}
	}

	return nil, nil
}

func (s *CharacterCreateScene) startGame() Scene {
	race := OrderedRaces[s.raceIdx]
	def := playerDefs[0] // デフォルトの冒険者定義を使用

	// レベルに応じた初期ステータスの計算（ランダム成長をシミュレート）
	str, wis, fai, vit, agi, luk := def.Str, def.Wis, def.Fai, def.Vit, def.Agi, def.Luk
	hp := def.HP
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// レベル1は成長なし。2以上から成長
	for l := 1; l < s.level; l++ {
		count := rng.Intn(3) + 1
		for i := 0; i < count; i++ {
			stat := rng.Intn(6)
			switch stat {
			case 0:
				str++
			case 1:
				wis++
			case 2:
				fai++
			case 3:
				vit++
				hp += 2
			case 4:
				agi++
			case 5:
				luk++
			}
		}
	}

	// プレイヤー生成
	p := NewUserPlayer("あなた", race, s.level, str, wis, fai, vit, agi, luk, def.Element, def.HP, def.MP)
	return newGameSceneWithState(p, nil, 0, 0, false, nil)
}

func (s *CharacterCreateScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "キャラクター作成", fontFace32, screenWidth/2, 80, color.RGBA{255, 220, 100, 255}, true)

	const startY = 200
	const rowH = 40

	// レベル
	var clrLevel color.Color = color.White
	if s.cursor == charCreateLevel {
		clrLevel = color.RGBA{255, 255, 0, 255}
	}
	drawText(screen, fmt.Sprintf("初期レベル: < %d >", s.level), fontFace14, screenWidth/2, startY, clrLevel, true)

	// 種族
	var clrRace color.Color = color.White
	if s.cursor == charCreateRace {
		clrRace = color.RGBA{255, 255, 0, 255}
	}
	raceName := RaceNames[OrderedRaces[s.raceIdx]]
	drawText(screen, fmt.Sprintf("種族: < %s >", raceName), fontFace14, screenWidth/2, startY+rowH, clrRace, true)

	// 開始
	var clrStart color.Color = color.White
	if s.cursor == charCreateStart {
		clrStart = color.RGBA{255, 255, 0, 255}
	}
	drawText(screen, "冒険を始める", fontFace20, screenWidth/2, startY+rowH*3, clrStart, true)

	drawText(screen, "W/S: 選択  A/D: 変更  Enter: 開始  Esc: 戻る", fontFace12, screenWidth/2, screenHeight-40, color.RGBA{150, 150, 150, 255}, true)
}
