package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"io"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/fonts/mplus1.ttf
var fontData []byte

//go:embed assets/sfx/Hit-1.wav
var sfxHitData []byte

//go:embed assets/sfx/stair_down.wav
var sfxStairDownData []byte

//go:embed assets/sfx/stair_up.wav
var sfxStairUpData []byte

//go:embed assets/sfx/Coin-3.wav
var sfxCoinData []byte

//go:embed assets/credits.txt
var creditsText string

var (
	fontFaceSource *text.GoTextFaceSource
	fontFace12     *text.GoTextFace
	fontFace14     *text.GoTextFace
	fontFace20     *text.GoTextFace
	fontFace32     *text.GoTextFace

	audioContext    *audio.Context
	sfxHitPCM       []byte
	sfxStairDownPCM []byte
	sfxStairUpPCM   []byte
	sfxCoinPCM      []byte
	sfxVolume       float64 = 0.5
)

const (
	screenWidth  = 640
	screenHeight = 480
	tileSize     = 16
	mapWidth     = 40
	mapHeight    = 25
)

// --- Scene インターフェース ---

type Scene interface {
	Update() (Scene, error)
	Draw(screen *ebiten.Image)
}

// --- App（Ebitenのトップレベル） ---

type App struct {
	scene Scene
}

func (a *App) Update() error {
	next, err := a.scene.Update()
	if err != nil {
		return err
	}
	if next != nil {
		a.scene = next
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	a.scene.Draw(screen)
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// --- タイトル画面 ---

type TitleScene struct{}

func (s *TitleScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return NewGameScene(), nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return &HelpScene{prev: s}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		return &OptionsScene{prev: s}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		return NewCreditScene(s), nil
	}
	return nil, nil
}

func (s *TitleScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "My Rogue", fontFace32, screenWidth/2, 120, color.RGBA{255, 220, 100, 255}, true)
	drawText(screen, "ダンジョンを探索し、宝を持ち帰れ", fontFace14, screenWidth/2, 180, color.RGBA{200, 200, 200, 255}, true)

	drawText(screen, "Enter / Space : ゲーム開始", fontFace14, screenWidth/2, 280, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "H : 操作説明", fontFace14, screenWidth/2, 310, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "O : オプション", fontFace14, screenWidth/2, 340, color.RGBA{255, 255, 255, 255}, true)
	drawText(screen, "C : クレジット", fontFace14, screenWidth/2, 370, color.RGBA{255, 255, 255, 255}, true)
}

// --- オプション画面 ---

type OptionsScene struct {
	prev Scene
}

func (s *OptionsScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyO) {
		return s.prev, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		sfxVolume -= 0.1
		if sfxVolume < 0 {
			sfxVolume = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		sfxVolume += 0.1
		if sfxVolume > 1.0 {
			sfxVolume = 1.0
		}
	}
	return nil, nil
}

func (s *OptionsScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "オプション", fontFace20, screenWidth/2, 80, color.RGBA{255, 220, 100, 255}, true)

	// SE音量スライダー
	const (
		barX     = 160
		barY     = 220
		barW     = 320
		barH     = 12
		knobSize = 18
	)
	drawText(screen, "SE 音量", fontFace14, screenWidth/2, 180, color.RGBA{220, 220, 220, 255}, true)

	// バー背景
	bar := ebiten.NewImage(barW, barH)
	bar.Fill(color.RGBA{60, 60, 80, 255})
	barOp := &ebiten.DrawImageOptions{}
	barOp.GeoM.Translate(float64(barX), float64(barY))
	screen.DrawImage(bar, barOp)

	// バー塗り
	filled := int(sfxVolume * float64(barW))
	if filled > 0 {
		fill := ebiten.NewImage(filled, barH)
		fill.Fill(color.RGBA{100, 200, 255, 255})
		fillOp := &ebiten.DrawImageOptions{}
		fillOp.GeoM.Translate(float64(barX), float64(barY))
		screen.DrawImage(fill, fillOp)
	}

	// ノブ
	knobX := barX + filled - knobSize/2
	knob := ebiten.NewImage(knobSize, knobSize)
	knob.Fill(color.RGBA{255, 255, 255, 255})
	knobOp := &ebiten.DrawImageOptions{}
	knobOp.GeoM.Translate(float64(knobX), float64(barY)-float64(knobSize-barH)/2)
	screen.DrawImage(knob, knobOp)

	pct := int(sfxVolume * 100)
	drawText(screen, fmt.Sprintf("%d%%", pct), fontFace14, screenWidth/2, barY+40, color.RGBA{255, 255, 255, 255}, true)

	drawText(screen, "<- / -> : 音量調整", fontFace12, screenWidth/2, 320, color.RGBA{180, 180, 180, 255}, true)
	drawText(screen, "Esc / O : 戻る", fontFace12, screenWidth/2, 345, color.RGBA{150, 150, 150, 255}, true)
}

// --- クレジット画面 ---

const (
	creditLineHeight = 28
	creditScrollSpeed = 1.0 // ピクセル/フレーム
)

type CreditScene struct {
	prev   Scene
	lines  []string
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

// --- ヘルプ画面 ---

type HelpScene struct {
	prev Scene // 戻り先（タイトルまたはゲーム画面）
}

func (s *HelpScene) Update() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return s.prev, nil
	}
	return nil, nil
}

func (s *HelpScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{10, 10, 30, 255})

	drawText(screen, "操作説明", fontFace20, screenWidth/2, 60, color.RGBA{255, 220, 100, 255}, true)

	lines := []string{
		"移動: 矢印キー / WASD",
		"攻撃: 敵のいる方向に移動（1ターン消費）",
		"",
		"目的: 宝（金色）を見つけて入り口（水色）に戻る",
		"    敵を避けるか倒しながら生き延びろ！",
		"",
		"色の意味:",
		"  灰色 = 壁    暗い灰 = 床",
		"  水色 = 階段（入り口）",
		"  金色 = 宝    紫色 = 敵",
		"  赤色 = プレイヤー",
		"  黄色 = プレイヤー（宝所持）",
		"",
		"HP: 20からスタート。敵に隣接されると攻撃される",
		"R: ゲームオーバー/クリア後にリスタート",
		"Esc: タイトルに戻る",
	}

	y := 120
	for _, line := range lines {
		if line == "" {
			y += 10
			continue
		}
		drawText(screen, line, fontFace14, 80, y, color.RGBA{220, 220, 220, 255}, false)
		y += 24
	}

	drawText(screen, "Esc / Backspace / H : 戻る", fontFace12, screenWidth/2, 430, color.RGBA{150, 150, 150, 255}, true)
}

// --- ゲーム画面 ---

type TileType int

const (
	Wall TileType = iota
	Floor
	Stairs     // 上り階段（フロア0の出口）
	Treasure
	StairsDown // 下り階段（次のフロアへ）
	StairsUp   // 上り階段（前のフロアへ、フロア1・2に配置）
)

const maxFloor = 3 // ダンジョンの深さ

type PlayState int

const (
	StatePlaying PlayState = iota
	StateWin
	StateDead
)

type Enemy struct {
	x, y int
}

type GameScene struct {
	worldMap    [mapWidth][mapHeight]TileType
	explored    [mapWidth][mapHeight]bool
	playerX     int
	playerY     int
	playerHP    int
	hasTreasure bool
	playState   PlayState
	message     string
	turnCount   int
	enemies     []Enemy
	rooms       []Room
	maxEnemies  int

	floor       int
}

func NewGameScene() *GameScene {
	return newGameSceneWithState(20, 0, 0, false, false)
}

func newGameSceneWithState(hp, turnCount, floor int, hasTreasure bool, fromBelow bool) *GameScene {
	g := &GameScene{playerHP: hp, turnCount: turnCount, floor: floor, hasTreasure: hasTreasure}
	g.generateMap()
	// 上り階段で戻った場合、プレイヤーを下り階段の位置に配置
	if fromBelow {
		for x := 0; x < mapWidth; x++ {
			for y := 0; y < mapHeight; y++ {
				if g.worldMap[x][y] == StairsDown {
					g.playerX = x
					g.playerY = y
				}
			}
		}
	}
	g.updateExplored()
	g.spawnInitialEnemies()
	switch {
	case floor == 0:
		g.message = fmt.Sprintf("フロア %d / %d。下り階段を探せ！", floor+1, maxFloor)
	case floor == maxFloor-1:
		g.message = fmt.Sprintf("フロア %d / %d。宝を見つけて戻れ！", floor+1, maxFloor)
	default:
		g.message = fmt.Sprintf("フロア %d / %d。さらに深く潜れ！", floor+1, maxFloor)
	}
	return g
}

type Room struct {
	x, y, w, h int
}

func (r Room) centerX() int { return r.x + r.w/2 }
func (r Room) centerY() int { return r.y + r.h/2 }

func (g *GameScene) generateMap() {
	rand.Seed(time.Now().UnixNano())

	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.worldMap[x][y] = Wall
		}
	}

	for i := 0; i < 5; i++ {
		w := rand.Intn(6) + 4
		h := rand.Intn(6) + 4
		x := rand.Intn(mapWidth-w-2) + 1
		y := rand.Intn(mapHeight-h-2) + 1

		for rx := x; rx < x+w; rx++ {
			for ry := y; ry < y+h; ry++ {
				g.worldMap[rx][ry] = Floor
			}
		}
		g.rooms = append(g.rooms, Room{x, y, w, h})
	}

	for i := 1; i < len(g.rooms); i++ {
		g.digCorridor(g.rooms[i-1].centerX(), g.rooms[i-1].centerY(), g.rooms[i].centerX(), g.rooms[i].centerY())
	}

	if len(g.rooms) > 0 {
		g.playerX = g.rooms[0].centerX()
		g.playerY = g.rooms[0].centerY()
		// フロア0のみ上り出口（Stairs）を配置、それ以外は上り階段（StairsUp）
		if g.floor == 0 {
			g.worldMap[g.playerX][g.playerY] = Stairs
		} else {
			g.worldMap[g.playerX][g.playerY] = StairsUp
		}
	}

	// 最下層のみ宝を配置
	if g.floor == maxFloor-1 {
		for {
			tx := rand.Intn(mapWidth)
			ty := rand.Intn(mapHeight)
			if g.worldMap[tx][ty] == Floor {
				g.worldMap[tx][ty] = Treasure
				break
			}
		}
	}

	// 最下層以外は下り階段を配置（上り階段のある rooms[0] 以外の部屋に限定）
	if g.floor < maxFloor-1 && len(g.rooms) >= 2 {
		for attempt := 0; attempt < 200; attempt++ {
			// rooms[1] 以降からランダムに選ぶ
			roomIdx := 1 + rand.Intn(len(g.rooms)-1)
			r := g.rooms[roomIdx]
			dx := r.x + rand.Intn(r.w)
			dy := r.y + rand.Intn(r.h)
			if g.worldMap[dx][dy] == Floor {
				g.worldMap[dx][dy] = StairsDown
				break
			}
		}
	}

	g.maxEnemies = len(g.rooms)
}

func (g *GameScene) digCorridor(x1, y1, x2, y2 int) {
	if rand.Intn(2) == 0 {
		g.digHorizontal(x1, x2, y1)
		g.digVertical(y1, y2, x2)
	} else {
		g.digVertical(y1, y2, x1)
		g.digHorizontal(x1, x2, y2)
	}
}

func (g *GameScene) digHorizontal(x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if g.worldMap[x][y] == Wall {
			g.worldMap[x][y] = Floor
		}
	}
}

func (g *GameScene) digVertical(y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if g.worldMap[x][y] == Wall {
			g.worldMap[x][y] = Floor
		}
	}
}

// プレイヤーがいる部屋のインデックスを返す（通路上なら-1）
func (g *GameScene) playerRoom() int {
	for i, r := range g.rooms {
		if g.playerX >= r.x && g.playerX < r.x+r.w && g.playerY >= r.y && g.playerY < r.y+r.h {
			return i
		}
	}
	return -1
}

// 指定座標が階段または宝の周囲3マス以内か
func (g *GameScene) nearSpecialTile(x, y int) bool {
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < mapWidth && ny >= 0 && ny < mapHeight {
				t := g.worldMap[nx][ny]
				if t == Stairs || t == StairsUp || t == StairsDown || t == Treasure {
					return true
				}
			}
		}
	}
	return false
}

// 初期敵配置
func (g *GameScene) spawnInitialEnemies() {
	playerRoomIdx := g.playerRoom()
	count := len(g.rooms) / 2
	for i := 0; i < count; i++ {
		g.trySpawnEnemy(playerRoomIdx)
	}
}

// 毎ターン敵の追加生成を試みる
func (g *GameScene) trySpawnEnemyPerTurn() {
	if len(g.enemies) >= g.maxEnemies {
		return
	}
	// 一定確率で生成（毎ターン30%）
	if rand.Intn(100) >= 30 {
		return
	}
	g.trySpawnEnemy(g.playerRoom())
}

// プレイヤーと同じ部屋以外の床にランダムに1体生成
func (g *GameScene) trySpawnEnemy(playerRoomIdx int) {
	for attempt := 0; attempt < 50; attempt++ {
		roomIdx := rand.Intn(len(g.rooms))
		if roomIdx == playerRoomIdx {
			continue
		}
		r := g.rooms[roomIdx]
		ex := r.x + rand.Intn(r.w)
		ey := r.y + rand.Intn(r.h)
		if g.worldMap[ex][ey] != Floor {
			continue
		}
		if g.nearSpecialTile(ex, ey) {
			continue
		}
		if g.isEnemyAt(ex, ey) {
			continue
		}
		g.enemies = append(g.enemies, Enemy{x: ex, y: ey})
		return
	}
}

func (g *GameScene) isEnemyAt(x, y int) bool {
	for _, e := range g.enemies {
		if e.x == x && e.y == y {
			return true
		}
	}
	return false
}

// プレイヤーが敵を攻撃（移動先の敵を倒す）
func (g *GameScene) attackEnemy(x, y int) {
	for i, e := range g.enemies {
		if e.x == x && e.y == y {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
			g.message = "敵を倒した！"
			playSFXHit()
			return
		}
	}
}

// 敵の移動（プレイヤーに向かって1マス移動、隣接時は攻撃）
func (g *GameScene) moveEnemies() {
	for i := range g.enemies {
		e := &g.enemies[i]

		// プレイヤーに隣接しているなら攻撃
		if abs(e.x-g.playerX)+abs(e.y-g.playerY) == 1 {
			g.playerHP--
			playSFXHit()
			if g.playerHP <= 0 {
				g.playState = StateDead
				g.message = "力尽きた..."
				return
			}
			g.message = fmt.Sprintf("敵に攻撃された！ HP: %d", g.playerHP)
			continue
		}

		dx, dy := 0, 0
		if e.x < g.playerX {
			dx = 1
		} else if e.x > g.playerX {
			dx = -1
		}
		if e.y < g.playerY {
			dy = 1
		} else if e.y > g.playerY {
			dy = -1
		}

		// 斜め移動はしない。ランダムで水平か垂直を選ぶ
		if dx != 0 && dy != 0 {
			if rand.Intn(2) == 0 {
				dy = 0
			} else {
				dx = 0
			}
		}

		nx, ny := e.x+dx, e.y+dy
		if nx >= 0 && nx < mapWidth && ny >= 0 && ny < mapHeight && g.worldMap[nx][ny] != Wall {
			if !g.isEnemyAt(nx, ny) && !(nx == g.playerX && ny == g.playerY) {
				e.x = nx
				e.y = ny
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *GameScene) Update() (Scene, error) {
	if g.playState == StateWin || g.playState == StateDead {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			return NewGameScene(), nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return &TitleScene{}, nil
		}
		return nil, nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return &TitleScene{}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return &HelpScene{prev: g}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		return &OptionsScene{prev: g}, nil
	}

	dx, dy := 0, 0
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		dx = -1
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		dx = 1
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		dy = -1
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		dy = 1
	}

	if dx != 0 || dy != 0 {
		newX, newY := g.playerX+dx, g.playerY+dy
		if newX >= 0 && newX < mapWidth && newY >= 0 && newY < mapHeight {
			if g.worldMap[newX][newY] != Wall {
				g.turnCount++
				if g.isEnemyAt(newX, newY) {
					// 攻撃（移動しない）
					g.attackEnemy(newX, newY)
				} else {
					// 移動
					g.playerX = newX
					g.playerY = newY
					g.updateExplored()
					if next, err := g.checkTile(); next != nil || err != nil {
						return next, err
					}
				}
				// 敵のターン
				g.moveEnemies()
				if g.playState == StateDead {
					return nil, nil
				}
				g.trySpawnEnemyPerTurn()
			}
		}
	}

	return nil, nil
}

func (g *GameScene) updateExplored() {
	radius := 4
	for x := g.playerX - radius; x <= g.playerX+radius; x++ {
		for y := g.playerY - radius; y <= g.playerY+radius; y++ {
			if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
				g.explored[x][y] = true
			}
		}
	}
}

func (g *GameScene) checkTile() (Scene, error) {
	tile := g.worldMap[g.playerX][g.playerY]
	switch tile {
	case Treasure:
		g.hasTreasure = true
		g.worldMap[g.playerX][g.playerY] = Floor
		g.message = "宝を手に入れた！上の階段まで戻れ！"
		playSFX(sfxCoinPCM)
	case Stairs: // フロア0の出口：宝持参でクリア
		if g.hasTreasure {
			g.playState = StateWin
			g.message = fmt.Sprintf("脱出成功！ スコア: %d（ターン: %d / HP: %d）", g.calcScore(), g.turnCount, g.playerHP)
			playSFX(sfxStairUpPCM)
		} else {
			g.message = "宝がない！まだ帰れない。"
		}
	case StairsUp: // フロア1・2：1つ上のフロアへ（下り階段位置にスポーン）
		playSFX(sfxStairUpPCM)
		return newGameSceneWithState(g.playerHP, g.turnCount, g.floor-1, g.hasTreasure, true), nil
	case StairsDown: // 下のフロアへ（上り階段位置にスポーン）
		playSFX(sfxStairDownPCM)
		return newGameSceneWithState(g.playerHP, g.turnCount, g.floor+1, g.hasTreasure, false), nil
	}
	return nil, nil
}

// スコア算出: (残りHP * 100 - ターン数) * maxFloor（最低0）
func (g *GameScene) calcScore() int {
	base := g.playerHP*100 - g.turnCount
	if base < 0 {
		base = 0
	}
	return base * maxFloor
}

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

	// UIメッセージ
	msgY := mapHeight*tileSize + 8
	drawText(screen, fmt.Sprintf("フロア: %d  ターン: %d  HP: %d", g.floor+1, g.turnCount, g.playerHP), fontFace12, 8, msgY, color.RGBA{200, 200, 200, 255}, false)
	drawText(screen, g.message, fontFace14, 8, msgY+20, color.RGBA{255, 255, 255, 255}, false)

	if g.playState == StateWin {
		drawText(screen, "R: 再挑戦 / Esc: タイトルへ", fontFace12, 8, msgY+44, color.RGBA{150, 255, 150, 255}, false)
	} else if g.playState == StateDead {
		drawText(screen, "R: 再挑戦 / Esc: タイトルへ", fontFace12, 8, msgY+44, color.RGBA{255, 100, 100, 255}, false)
	} else {
		drawText(screen, "H: ヘルプ / O: オプション / Esc: タイトル", fontFace12, screenWidth-290, msgY, color.RGBA{100, 100, 100, 255}, false)
	}
}

// --- テキスト描画ヘルパー ---

func drawText(screen *ebiten.Image, str string, face *text.GoTextFace, x, y int, clr color.Color, center bool) {
	op := &text.DrawOptions{}
	if center {
		op.PrimaryAlign = text.AlignCenter
	}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, str, face, op)
}

// --- 初期化・メイン ---

func playSFX(pcm []byte) {
	if pcm == nil || sfxVolume <= 0 {
		return
	}
	p := audioContext.NewPlayerFromBytes(pcm)
	p.SetVolume(sfxVolume)
	p.Play()
}

func playSFXHit() { playSFX(sfxHitPCM) }

func init() {
	var err error
	fontFaceSource, err = text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Fatal(err)
	}
	fontFace12 = &text.GoTextFace{Source: fontFaceSource, Size: 12}
	fontFace14 = &text.GoTextFace{Source: fontFaceSource, Size: 14}
	fontFace20 = &text.GoTextFace{Source: fontFaceSource, Size: 20}
	fontFace32 = &text.GoTextFace{Source: fontFaceSource, Size: 32}

	audioContext = audio.NewContext(44100)

	decodeSFX := func(data []byte) []byte {
		s, err := wav.DecodeWithSampleRate(44100, bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
		pcm, err := io.ReadAll(s)
		if err != nil {
			log.Fatal(err)
		}
		return pcm
	}

	sfxHitPCM = decodeSFX(sfxHitData)
	sfxStairDownPCM = decodeSFX(sfxStairDownData)
	sfxStairUpPCM = decodeSFX(sfxStairUpData)
	sfxCoinPCM = decodeSFX(sfxCoinData)
}

func main() {
	app := &App{scene: &TitleScene{}}
	ebiten.SetWindowTitle("My Rogue")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
