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
	ID                           int64
	X, Y                         int
	Dir                          Dir
	HP                           int
	MaxHP                        int
	MP                           int
	MaxMP                        int
	Level                        int
	XP                           int
	XPToNext                     int
	Str, Wis, Fai, Vit, Agi, Luk int
	Element                      Element
	Race                         Race
	AttackAnim                   int // 攻撃アニメ残りフレーム数
	DamageAnim                   int // 被ダメージアニメ残りフレーム数
	Inventory                    []InventoryEntry
	Relations                    map[int64]int // ID -> 感情値
}

func (a *Actor) GetID() int64 { return a.ID }
func (a *Actor) GetStats() Stats {
	return Stats{
		Attack:  1 + a.Str + a.equippedAtk(),
		Defense: a.equippedDef() + a.Vit,
		Element: a.Element,
	}
}
func (a *Actor) GetName() string {
	return "???"
}
func (a *Actor) GetRace() Race { return a.Race }
func (a *Actor) GetLevel() int { return a.Level }
func (a *Actor) ApplyDamage(dmg int) {
	a.HP -= dmg
	a.DamageAnim = damageAnimFrames
}
func (a *Actor) UpdateRelation(targetID int64, delta int) {
	if a.Relations == nil {
		a.Relations = make(map[int64]int)
	}
	a.Relations[targetID] += delta
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
	worldMap       [mapWidth][mapHeight]TileType
	explored       [mapWidth][mapHeight]bool // 一度でも視界に入ったタイル（静的要素を表示）
	visible        [mapWidth][mapHeight]bool // 現在視界に入っているタイル（動的要素も表示）
	mapSeed        int64                     // このフロアのマップ生成シード
	Player         Actor
	playState      PlayState
	turnState      TurnState
	message        string
	messageLog     []string
	projectiles    []Projectile
	turnCount      int
	enemies        []Enemy
	activeEnemyIdx int // 現在行動中の敵のインデックス
	rooms          []Room
	maxEnemies     int
	floor          int
	Combat         *CombatManager

	menuOpen   bool
	menuCursor int
	menuItems  []actionItem

	mapItems []MapItem

	frame           int     // 毎フレームインクリメント（アニメ用）
	AnimSpeed       float64 // アニメーション速度
	animAccumulator float64 // 速度調整用蓄積

	confirmQuit bool // タイトルへ戻る確認ダイアログ表示中
}

func (g *GameScene) HasTreasure() bool {
	hasTreasure := false
	for _, entry := range g.Player.Inventory {
		if entry.kind == itemIDMap["treasure"] {
			hasTreasure = true
			break
		}
	}
	return hasTreasure
}

func (g *GameScene) GetID() int64 { return g.Player.ID }
func (g *GameScene) GetStats() Stats {
	return g.Player.GetStats()
}

func (g *GameScene) ApplyDamage(dmg int) {
	g.Player.ApplyDamage(dmg)
}

func (g *GameScene) GetName() string {
	return "あなた"
}

func (g *GameScene) GetRace() Race {
	return g.Player.Race
}

func (g *GameScene) GetLevel() int {
	return g.Player.Level
}

func (g *GameScene) UpdateRelation(targetID int64, delta int) {
	g.Player.UpdateRelation(targetID, delta)
}

func (g *GameScene) GetRelation(target Battler) RelationState {
	if g.Player.Race == target.GetRace() {
		return RelationFriendly
	}
	score := g.Player.Relations[target.GetID()]
	if score <= -20 {
		return RelationHostile
	}
	if score >= 50 {
		return RelationFriendly
	}
	return RelationNeutral
}

func (g *GameScene) ConsumeDurability(gs *GameScene, et EquipType) {
	if et == EquipNone {
		return
	}
	for i := 0; i < len(g.Player.Inventory); i++ {
		entry := &g.Player.Inventory[i]
		if entry.Equipped && itemDefs[entry.kind].equipType == et {
			def := &itemDefs[entry.kind]
			if def.durability < 0 {
				return // 壊れない
			}
			entry.Durability--
			if entry.Durability <= 0 {
				g.pushMessage(def.name + "は壊れてしまった！")
				g.Player.Inventory = append(g.Player.Inventory[:i], g.Player.Inventory[i+1:]...)
			}
			return
		}
	}
}

func (g *GameScene) dropChest(x, y int, inventory []InventoryEntry) {
	if len(inventory) == 0 {
		return
	}
	for i := range inventory {
		inventory[i].Equipped = false
	}
	for i := range g.mapItems {
		it := &g.mapItems[i]
		if it.X == x && it.Y == y {
			it.Inventory = append(it.Inventory, inventory...)
			g.pushMessage("足元の宝箱にアイテムが追加された。")
			return
		}
	}
	g.mapItems = append(g.mapItems, MapItem{
		X: x, Y: y,
		Inventory: inventory,
	})
	g.pushMessage("宝箱がドロップされた！")
}

func NewGameScene() *GameScene {
	return newGameSceneWithState(20, 0, 0, false, nil, nil, 1, 0, 10, 1, 1, 1, 1, 1, 1, RaceHuman)
}

func newGameSceneWithState(hp, turnCount, floor int, fromBelow bool, inventory []InventoryEntry, log []string, level, xp, nextXP, str, wis, fai, vit, agi, luk int, race Race) *GameScene {
	g := &GameScene{
		Player: Actor{
			ID:        1, // プレイヤーは固定ID 1
			HP:        hp,
			MaxHP:     playerMaxHP + vit*2,
			Dir:       DirDown,
			Element:   ElementNone,
			Race:      race,
			Inventory: inventory,
			Level:     level,
			XP:        xp,
			XPToNext:  nextXP,
			Str:       str,
			Wis:       wis,
			Fai:       fai,
			Vit:       vit,
			Agi:       agi,
			Luk:       luk,
			Relations: make(map[int64]int),
		},
		turnCount:  turnCount,
		floor:      floor,
		messageLog: log,
		Combat:     &CombatManager{},
		AnimSpeed:  1.0,
	}
	g.Player.MaxMP = 5 + g.Player.Wis*5
	g.Player.MP = 5 + g.Player.Wis*5

	g.generateMap()
	if fromBelow {
		for x := 0; x < mapWidth; x++ {
			for y := 0; y < mapHeight; y++ {
				if g.worldMap[x][y] == StairsDown {
					g.Player.X, g.Player.Y = x, y
				}
			}
		}
	}
	g.updateVisibility()
	g.spawnInitialEnemies()
	return g
}

func (g *GameScene) actorGainXP(a *Actor, amount int) {
	if a.Level >= 100 {
		return
	}
	a.XP += amount
	if a.ID == g.Player.ID {
		g.pushMessage(fmt.Sprintf("%d の経験値を獲得！", amount))
	}
	for a.XP >= a.XPToNext && a.Level < 100 {
		a.XP -= a.XPToNext
		a.Level++
		a.XPToNext = int(10 * math.Pow(float64(a.Level), 1.5))
		if a.ID == g.Player.ID {
			g.pushMessage(fmt.Sprintf("レベル %d に上がった！", a.Level))
		}
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + a.ID))
		g.rollStatsUp(a, rng)
	}
}

func (g *GameScene) rollStatsUp(a *Actor, rng *rand.Rand) {
	count := rng.Intn(3) + 1
	for i := 0; i < count; i++ {
		stat := rng.Intn(6)
		switch stat {
		case 0:
			a.Str++
		case 1:
			a.Wis++
		case 2:
			a.Fai++
		case 3:
			a.Vit++
			a.HP += 2
			a.MaxHP += 2
		case 4:
			a.Agi++
		case 5:
			a.Luk++
		}
	}
}

func (g *GameScene) tryDodge(acc int) bool {
	dodgeChance := (g.Player.Agi*2 + g.Player.Luk/2) + (100 - acc)
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
	return len(g.projectiles) > 0
}

func (g *GameScene) updateAnimations() {
	g.animAccumulator += g.AnimSpeed
	for g.animAccumulator >= 1.0 {
		g.animAccumulator -= 1.0

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

		newProjectiles := []Projectile{}
		for _, p := range g.projectiles {
			p.Frame++
			if p.Frame < p.TotalFrames {
				newProjectiles = append(newProjectiles, p)
			}
		}
		g.projectiles = newProjectiles
	}
}

func (g *GameScene) Update() (Scene, error) {
	g.frame++

	if next, err := g.handleDebugKeys(); next != nil || err != nil {
		return next, err
	}
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
		if g.Player.MP < g.Player.MaxMP {
			g.Player.MP++
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

		if g.HasTreasure() {
			g.playState = StateWin
			g.pushMessage(fmt.Sprintf("脱出成功！ スコア: %d（ターン: %d / HP: %d）", g.calcScore(), g.turnCount, g.Player.HP))
			playSFX(sfxStairUpPCM)
		} else {
			g.pushMessage("宝がない！まだ帰れない。")
		}
	case StairsUp:
		playSFX(sfxStairUpPCM)
		return newGameSceneWithState(g.Player.HP, g.turnCount, g.floor-1, true, g.Player.Inventory, g.messageLog, g.Player.Level, g.Player.XP, g.Player.XPToNext, g.Player.Str, g.Player.Wis, g.Player.Fai, g.Player.Vit, g.Player.Agi, g.Player.Luk, g.Player.Race), nil
	case StairsDown:
		playSFX(sfxStairDownPCM)
		return newGameSceneWithState(g.Player.HP, g.turnCount, g.floor+1, false, g.Player.Inventory, g.messageLog, g.Player.Level, g.Player.XP, g.Player.XPToNext, g.Player.Str, g.Player.Wis, g.Player.Fai, g.Player.Vit, g.Player.Agi, g.Player.Luk, g.Player.Race), nil
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
