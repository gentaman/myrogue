package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PlayState はゲームの状態を表す
type PlayState int

const (
	StatePlaying PlayState = iota
	StateWin
	StateDead
)

// GameScene はゲームプレイ画面を表す
type GameScene struct {
	worldMap    [mapWidth][mapHeight]TileType
	explored    [mapWidth][mapHeight]bool // 一度でも視界に入ったタイル（静的要素を表示）
	visible     [mapWidth][mapHeight]bool // 現在視界に入っているタイル（動的要素も表示）
	mapSeed     int64                     // このフロアのマップ生成シード
	playerX     int
	playerY     int
	playerHP    int
	MP          int
	MaxMP       int
	Level       int
	XP          int
	XPToNext    int
	Str, Wis, Fai, Vit, Agi, Luk int
	hasTreasure bool
	playState   PlayState
	message     string
	messageLog  []string
	turnCount   int
	enemies     []Enemy
	rooms       []Room
	maxEnemies  int
	floor       int
	playerDir   Dir

	menuOpen   bool
	menuCursor int
	menuItems  []actionItem

	mapItems  []MapItem
	inventory []InventoryEntry

	frame            int // 毎フレームインクリメント（アニメ用）
	playerAttackAnim int // プレイヤー攻撃アニメ残りフレーム数

	confirmQuit bool // タイトルへ戻る確認ダイアログ表示中
}

func NewGameScene() *GameScene {
	// 初期ステータス（ベースはすべて1）
	return newGameSceneWithState(20, 0, 0, false, false, nil, nil, 1, 0, 10, 1, 1, 1, 1, 1, 1)
}

func newGameSceneWithState(hp, turnCount, floor int, hasTreasure bool, fromBelow bool, inventory []InventoryEntry, log []string, level, xp, nextXP, str, wis, fai, vit, agi, luk int) *GameScene {
	g := &GameScene{
		playerHP:    hp,
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
					g.playerX = x
					g.playerY = y
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
		case 0: g.Str++; g.pushMessage("力が少し上がった！")
		case 1: g.Wis++; g.pushMessage("知恵が少し上がった！")
		case 2: g.Fai++; g.pushMessage("信仰心が少し上がった！")
		case 3: 
			g.Vit++
			g.playerHP += 2
			g.pushMessage("生命力が少し上がった！")
		case 4: g.Agi++; g.pushMessage("素早さが少し上がった！")
		case 5: g.Luk++; g.pushMessage("運が少し上がった！")
		}
	}
}

func (g *GameScene) tryDodge(acc int) bool {
	// evasionChance = Agi * 2 + Luk / 2 (max 80%)
	// 命中率(acc)で補正する (例: 命中が低いほど回避しやすい)
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

func (g *GameScene) Update() (Scene, error) {
	g.frame++
	// アニメカウンタを進める
	if g.playerAttackAnim > 0 {
		g.playerAttackAnim--
	}
	for i := range g.enemies {
		if g.enemies[i].attackAnim > 0 {
			g.enemies[i].attackAnim--
		}
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

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if g.menuOpen {
			g.menuOpen = false
			return nil, nil
		}
		g.confirmQuit = true
		return nil, nil
	}
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

	// アクションメニュー
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
			return g.execMenuItem()
		}
		return nil, nil
	}

	var newDir Dir
	var dirPressed bool
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
		newDir, dirPressed = DirLeft, true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
		newDir, dirPressed = DirRight, true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		newDir, dirPressed = DirUp, true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		newDir, dirPressed = DirDown, true
	}

	if dirPressed {
		dx, dy := newDir.delta()
		newX, newY := g.playerX+dx, g.playerY+dy

		// 向きは常に更新
		g.playerDir = newDir

		if newX >= 0 && newX < mapWidth && newY >= 0 && newY < mapHeight {
			if g.worldMap[newX][newY] != Wall {
				g.turnCount++
				if g.isEnemyAt(newX, newY) {
					// 正面の敵を攻撃（移動しない）
					g.playerAttackAnim = attackAnimFrames
					g.attackEnemy(newX, newY)
				} else {
					// 移動
					g.playerX = newX
					g.playerY = newY
					g.updateVisibility()
				}
				// 敵のターン
				g.moveEnemies()
				if g.playState == StateDead {
					return nil, nil
				}
				g.trySpawnEnemyPerTurn()
			} else {
				// 壁でも向きだけ変える（ターン消費なし）
			}
		}
	}

	return nil, nil
}

func (g *GameScene) updateVisibility() {
	// visible をリセット
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.visible[x][y] = false
		}
	}

	roomIdx := g.playerRoom()
	if roomIdx >= 0 {
		// 部屋内：部屋全体 + 部屋の外周1マスを visible に
		r := g.rooms[roomIdx]
		for x := r.x - 1; x < r.x+r.w+1; x++ {
			for y := r.y - 1; y < r.y+r.h+1; y++ {
				if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
					g.visible[x][y] = true
				}
			}
		}
	} else {
		// 通路：自マス + 周囲1マス（8方向）
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				x := g.playerX + dx
				y := g.playerY + dy
				if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
					g.visible[x][y] = true
				}
			}
		}
	}

	// visible になったタイルは explored にも追加
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			if g.visible[x][y] {
				g.explored[x][y] = true
			}
		}
	}
}

func (g *GameScene) checkTile() (Scene, error) {
	tile := g.worldMap[g.playerX][g.playerY]
	switch tile {
	case Stairs: // フロア0の出口：宝持参でクリア
		if g.hasTreasure {
			g.playState = StateWin
			g.pushMessage(fmt.Sprintf("脱出成功！ スコア: %d（ターン: %d / HP: %d）", g.calcScore(), g.turnCount, g.playerHP))
			playSFX(sfxStairUpPCM)
		} else {
			g.pushMessage("宝がない！まだ帰れない。")
		}
	case StairsUp: // フロア1・2：1つ上のフロアへ（下り階段位置にスポーン）
		playSFX(sfxStairUpPCM)
		return newGameSceneWithState(g.playerHP, g.turnCount, g.floor-1, g.hasTreasure, true, g.inventory, g.messageLog, g.Level, g.XP, g.XPToNext, g.Str, g.Wis, g.Fai, g.Vit, g.Agi, g.Luk), nil
	case StairsDown: // 下のフロアへ（上り階段位置にスポーン）
		playSFX(sfxStairDownPCM)
		return newGameSceneWithState(g.playerHP, g.turnCount, g.floor+1, g.hasTreasure, false, g.inventory, g.messageLog, g.Level, g.XP, g.XPToNext, g.Str, g.Wis, g.Fai, g.Vit, g.Agi, g.Luk), nil
	}
	return nil, nil
}

// スコア算出: (残りHP * 100 - ターン数) * 階層数（最低0）
func (g *GameScene) calcScore() int {
	base := g.playerHP*100 - g.turnCount
	if base < 0 {
		base = 0
	}
	return base * len(floorDefs)
}
