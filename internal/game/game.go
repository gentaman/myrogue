package game

import (
	"fmt"

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
	worldMap [mapWidth][mapHeight]TileType
	explored [mapWidth][mapHeight]bool // 一度でも視界に入ったタイル（静的要素を表示）
	visible  [mapWidth][mapHeight]bool // 現在視界に入っているタイル（動的要素も表示）
	mapSeed  int64                     // このフロアのマップ生成シード
	playerX  int
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
	return newGameSceneWithState(20, 0, 0, false, false, nil)
}

func newGameSceneWithState(hp, turnCount, floor int, hasTreasure bool, fromBelow bool, inventory []InventoryEntry) *GameScene {
	g := &GameScene{playerHP: hp, turnCount: turnCount, floor: floor, hasTreasure: hasTreasure, inventory: inventory}
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
		g.message = fmt.Sprintf("フロア %d / %d。下り階段を探せ！", floor+1, maxFloor)
	case floor == maxFloor-1:
		g.message = fmt.Sprintf("フロア %d / %d。宝を見つけて戻れ！", floor+1, maxFloor)
	default:
		g.message = fmt.Sprintf("フロア %d / %d。さらに深く潜れ！", floor+1, maxFloor)
	}
	return g
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
					g.pickupItem()
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
		return newGameSceneWithState(g.playerHP, g.turnCount, g.floor-1, g.hasTreasure, true, g.inventory), nil
	case StairsDown: // 下のフロアへ（上り階段位置にスポーン）
		playSFX(sfxStairDownPCM)
		return newGameSceneWithState(g.playerHP, g.turnCount, g.floor+1, g.hasTreasure, false, g.inventory), nil
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
