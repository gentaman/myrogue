package game

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var (
	//go:embed assets/images/player.png
	playerPNG []byte

	playerImage *ebiten.Image
)

func init() {
	// プレイヤー画像ロード
	img, _, err := image.Decode(bytes.NewReader(playerPNG))
	if err != nil {
		panic(fmt.Sprintf("failed to decode player.png: %v", err))
	}
	playerImage = ebiten.NewImageFromImage(img)
}

// PlayState はゲームの状態を表す
type PlayState int

const (
	StatePlaying PlayState = iota
	StateWin
	StateDead
)

// TurnState はターンの進行状態を表す
type TurnState int

const (
	TurnPlayerInput TurnState = iota
	TurnEnemyAct
)

// Actor はプレイヤーや敵の共通データ
type Actor struct {
	X, Y       int
	Dir        Dir
	HP         int
	MaxHP      int
	AttackAnim int // 攻撃アニメ残りフレーム数
	DamageAnim int // 被ダメージアニメ残りフレーム数
}

// Projectile は遠隔攻撃のアニメーション体
type Projectile struct {
	StartX, StartY float64
	EndX, EndY     float64
	Frame          int
	TotalFrames    int
	Color          color.RGBA
}

// GameScene はゲームプレイ画面を表す
type GameScene struct {
	worldMap                     [mapWidth][mapHeight]TileType
	explored                     [mapWidth][mapHeight]bool // 一度でも視界に入ったタイル（静的要素を表示）
	visible                      [mapWidth][mapHeight]bool // 現在視界に入っているタイル（動的要素も表示）
	mapSeed                      int64                     // このフロアのマップ生成シード
	Player                       Actor
	MP                           int
	MaxMP                        int
	Level                        int
	XP                           int
	XPToNext                     int
	Str, Wis, Fai, Vit, Agi, Luk int
	hasTreasure                  bool
	playState                    PlayState
	turnState                    TurnState
	message                      string
	messageLog                   []string
	projectiles                  []Projectile
	turnCount                    int
	enemies                      []Enemy
	activeEnemyIdx               int // 現在行動中の敵のインデックス
	rooms                        []Room
	maxEnemies                   int
	floor                        int

	menuOpen   bool
	menuCursor int
	menuItems  []actionItem

	mapItems  []MapItem
	inventory []InventoryEntry

	frame int // 毎フレームインクリメント（アニメ用）

	confirmQuit bool // タイトルへ戻る確認ダイアログ表示中
}

func NewGameScene() *GameScene {
	// 初期ステータス（ベースはすべて1）
	return newGameSceneWithState(20, 0, 0, false, false, nil, nil, 1, 0, 10, 1, 1, 1, 1, 1, 1)
}

func newGameSceneWithState(hp, turnCount, floor int, hasTreasure bool, fromBelow bool, inventory []InventoryEntry, log []string, level, xp, nextXP, str, wis, fai, vit, agi, luk int) *GameScene {
	g := &GameScene{
		Player: Actor{
			HP:    hp,
			MaxHP: playerMaxHP + vit*2,
			Dir:   DirDown,
		},
		MaxMP:       5 + wis*5,
		MP:          5 + wis*5,
		turnCount:   turnCount,
		floor:       floor,
		hasTreasure: hasTreasure,
		inventory:   inventory,
		messageLog:  log,
		Level:       level,
		XP:          xp,
		XPToNext:    nextXP,
		Str:         str,
		Wis:         wis,
		Fai:         fai,
		Vit:         vit,
		Agi:         agi,
		Luk:         luk,
	}
	g.generateMap()
	// 上り階段で戻った場合、プレイヤーを下り階段の位置に配置
	if fromBelow {
		for x := 0; x < mapWidth; x++ {
			for y := 0; y < mapHeight; y++ {
				if g.worldMap[x][y] == StairsDown {
					g.Player.X = x
					g.Player.Y = y
				}
			}
		}
	}
	g.updateVisibility()
	g.spawnInitialEnemies()
	switch {
	case floor == 0:
		g.pushMessage(fmt.Sprintf("フロア %d / %d。下り階段を探せ！", floor+1, len(floorDefs)))
	case floor == len(floorDefs)-1:
		g.pushMessage(fmt.Sprintf("フロア %d / %d。宝を見つけて戻れ！", floor+1, len(floorDefs)))
	default:
		if hasTreasure {
			g.pushMessage(fmt.Sprintf("フロア %d / %d。入口に戻れ！", floor+1, len(floorDefs)))
		} else {
			g.pushMessage(fmt.Sprintf("フロア %d / %d。さらに深く潜れ！", floor+1, len(floorDefs)))
		}
	}
	return g
}

func (g *GameScene) gainXP(amount int) {
	if g.Level >= 100 {
		return
	}
	g.XP += amount
	g.pushMessage(fmt.Sprintf("%d の経験値を獲得！", amount))
	for g.XP >= g.XPToNext && g.Level < 100 {
		g.XP -= g.XPToNext
		g.Level++
		// 10 * (level ^ 1.5)
		g.XPToNext = int(10 * math.Pow(float64(g.Level), 1.5))
		g.pushMessage(fmt.Sprintf("レベル %d に上がった！", g.Level))

		// ステータス上昇
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		g.rollStatsUp(rng)
	}
}

func (g *GameScene) rollStatsUp(rng *rand.Rand) {
	// 1-3個のステータスが上昇
	count := rng.Intn(3) + 1
	for i := 0; i < count; i++ {
		stat := rng.Intn(6)
		switch stat {
		case 0:
			g.Str++
			g.pushMessage("力が少し上がった！")
		case 1:
			g.Wis++
			g.pushMessage("知恵が少し上がった！")
		case 2:
			g.Fai++
			g.pushMessage("信仰心が少し上がった！")
		case 3:
			g.Vit++
			g.Player.HP += 2
			g.Player.MaxHP += 2
			g.pushMessage("生命力が少し上がった！")
		case 4:
			g.Agi++
			g.pushMessage("素早さが少し上がった！")
		case 5:
			g.Luk++
			g.pushMessage("運が少し上がった！")
		}
	}
}

func (g *GameScene) tryDodge(acc int) bool {
	dodgeChance := (g.Agi*2 + g.Luk/2) + (100 - acc)
	if dodgeChance > 80 {
		dodgeChance = 80
	}
	return rand.Intn(100) < dodgeChance
}

func (g *GameScene) pushMessage(msg string) {
	g.message = msg
	g.messageLog = append(g.messageLog, msg)
	if len(g.messageLog) > 1000 {
		g.messageLog = g.messageLog[len(g.messageLog)-1000:]
	}
}

// カメラのオフセットを計算する
func (g *GameScene) cameraOffset() (float64, float64) {
	camX := float64(g.Player.X*tileSize - screenWidth/2 + tileSize/2)
	camY := float64(g.Player.Y*tileSize - gameplayAreaHeight/2 + tileSize/2)

	if camX < 0 {
		camX = 0
	}
	if camY < 0 {
		camY = 0
	}
	maxCamX := float64(mapWidth*tileSize - screenWidth)
	maxCamY := float64(mapHeight*tileSize - gameplayAreaHeight)
	if camX > maxCamX {
		camX = maxCamX
	}
	if camY > maxCamY {
		camY = maxCamY
	}
	return camX, camY
}

func (g *GameScene) isAnyAnimating() bool {
	if g.Player.AttackAnim > 0 || g.Player.DamageAnim > 0 {
		return true
	}
	for _, e := range g.enemies {
		if e.AttackAnim > 0 || e.DamageAnim > 0 {
			return true
		}
	}
	if len(g.projectiles) > 0 {
		return true
	}
	return false
}

func (g *GameScene) updateAnimations() {
	if g.Player.AttackAnim > 0 {
		g.Player.AttackAnim--
	}
	if g.Player.DamageAnim > 0 {
		g.Player.DamageAnim--
	}
	for i := range g.enemies {
		if g.enemies[i].AttackAnim > 0 {
			g.enemies[i].AttackAnim--
		}
		if g.enemies[i].DamageAnim > 0 {
			g.enemies[i].DamageAnim--
		}
	}
}

func (g *GameScene) Update() (Scene, error) {
	g.frame++

	newProjectiles := []Projectile{}
	for _, p := range g.projectiles {
		p.Frame++
		if p.Frame < p.TotalFrames {
			newProjectiles = append(newProjectiles, p)
		}
	}
	g.projectiles = newProjectiles

	if g.isAnyAnimating() {
		g.updateAnimations()
		return nil, nil
	}

	if g.playState == StateWin || g.playState == StateDead {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			return NewGameScene(), nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return &TitleScene{}, nil
		}
		return nil, nil
	}
	if g.confirmQuit {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
			g.confirmQuit = false
		} else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyY) {
			return &TitleScene{}, nil
		}
		return nil, nil
	}

	switch g.turnState {
	case TurnPlayerInput:
		return g.updatePlayerTurn()
	case TurnEnemyAct:
		g.updateEnemyTurns()
	}

	return nil, nil
}

func (g *GameScene) updatePlayerTurn() (Scene, error) {
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		return &HelpScene{prev: g}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		return &OptionsScene{prev: g}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		return &InventoryScene{game: g}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		return &LogScene{game: g}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		return &MapScene{game: g}, nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.menuOpen {
			g.menuOpen = false
			return nil, nil
		}
		g.confirmQuit = true
		return nil, nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyX) {
		g.menuOpen = !g.menuOpen
		if g.menuOpen {
			g.menuCursor = 0
			g.buildMenu()
		}
		return nil, nil
	}
	if g.menuOpen {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.menuCursor = (g.menuCursor - 1 + len(g.menuItems)) % len(g.menuItems)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.menuCursor = (g.menuCursor + 1) % len(g.menuItems)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			scene, err := g.execMenuItem()
			if scene != nil || err != nil {
				return scene, err
			}
			g.turnState = TurnEnemyAct
			g.activeEnemyIdx = 0
		}
		return nil, nil
	}

	var newDir Dir
	var dirPressed bool
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		newDir, dirPressed = DirLeft, true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		newDir, dirPressed = DirRight, true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		newDir, dirPressed = DirUp, true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		newDir, dirPressed = DirDown, true
	}

	if dirPressed {
		g.Player.Dir = newDir
		dx, dy := newDir.delta()
		nx, ny := g.Player.X+dx, g.Player.Y+dy
		if nx >= 0 && nx < mapWidth && ny >= 0 && ny < mapHeight {
			if g.worldMap[nx][ny] != Wall {
				g.turnCount++
				if g.isEnemyAt(nx, ny) {
					g.Player.AttackAnim = attackAnimFrames
					g.attackEnemy(nx, ny)
				} else {
					g.Player.X, g.Player.Y = nx, ny
					g.updateVisibility()
				}
				g.turnState = TurnEnemyAct
				g.activeEnemyIdx = 0
			}
		}
	}
	return nil, nil
}

func (g *GameScene) updateEnemyTurns() {
	if g.activeEnemyIdx >= len(g.enemies) {
		g.turnState = TurnPlayerInput
		if g.MP < g.MaxMP {
			g.MP++
		}
		g.trySpawnEnemyPerTurn()
		return
	}
	g.moveSingleEnemy(g.activeEnemyIdx)
	g.activeEnemyIdx++
}

func (g *GameScene) updateVisibility() {
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.visible[x][y] = false
		}
	}
	roomIdx := g.playerRoom()
	if roomIdx >= 0 {
		r := g.rooms[roomIdx]
		for x := r.x - 1; x < r.x+r.w+1; x++ {
			for y := r.y - 1; y < r.y+r.h+1; y++ {
				if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
					g.visible[x][y] = true
				}
			}
		}
	} else {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				x, y := g.Player.X+dx, g.Player.Y+dy
				if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
					g.visible[x][y] = true
				}
			}
		}
	}
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			if g.visible[x][y] {
				g.explored[x][y] = true
			}
		}
	}
}

func (g *GameScene) checkTile() (Scene, error) {
	tile := g.worldMap[g.Player.X][g.Player.Y]
	switch tile {
	case Stairs:
		if g.hasTreasure {
			g.playState = StateWin
			g.pushMessage(fmt.Sprintf("脱出成功！ スコア: %d（ターン: %d / HP: %d）", g.calcScore(), g.turnCount, g.Player.HP))
			playSFX(sfxStairUpPCM)
		} else {
			g.pushMessage("宝がない！まだ帰れない。")
		}
	case StairsUp:
		playSFX(sfxStairUpPCM)
		return newGameSceneWithState(g.Player.HP, g.turnCount, g.floor-1, g.hasTreasure, true, g.inventory, g.messageLog, g.Level, g.XP, g.XPToNext, g.Str, g.Wis, g.Fai, g.Vit, g.Agi, g.Luk), nil
	case StairsDown:
		playSFX(sfxStairDownPCM)
		return newGameSceneWithState(g.Player.HP, g.turnCount, g.floor+1, g.hasTreasure, false, g.inventory, g.messageLog, g.Level, g.XP, g.XPToNext, g.Str, g.Wis, g.Fai, g.Vit, g.Agi, g.Luk), nil
	}
	return nil, nil
}

func (g *GameScene) calcScore() int {
	base := g.Player.HP*100 - g.turnCount
	if base < 0 {
		base = 0
	}
	return base * len(floorDefs)
}
