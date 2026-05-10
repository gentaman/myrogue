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

const (
	MaxStr = 100
	MaxWis = 100
	MaxFai = 100
	MaxVit = 100
	MaxAgi = 100
	MaxLuk = 100
)

func (a *Actor) GetID() int64 { return a.ID }
func (a *Actor) GetStats() Stats {
	return Stats{
		PhysicalAttack:  a.Str + a.equippedPhyAtk(),
		PhysicalDefense: a.Vit + a.equippedPhyDef(),
		MagicalAttack:   a.Wis + a.equippedMagAtk(),
		MagicalDefense:  a.Fai + a.equippedMagDef(),
		BaseAccuracy:    100,
		Element:         a.Element,
	}
}

func (a *Actor) GetCombatType() CombatType {
	return CombatTypePhysical
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

func (a *Actor) GetRelation(target Battler) RelationState {
	if a.Race == target.GetRace() {
		return RelationFriendly
	}
	score := a.Relations[target.GetID()]
	if score <= -20 {
		return RelationHostile
	}
	if score >= 50 {
		return RelationFriendly
	}
	return RelationNeutral
}

func (a *Actor) ConsumeDurability(bus *EventBus, et EquipType) {
	if et == EquipNone {
		return
	}
	for i := 0; i < len(a.Inventory); i++ {
		entry := a.Inventory[i]
		if entry.Equipped && itemDefs[entry.kind].equipType == et {
			def := &itemDefs[entry.kind]
			if def.durability < 0 {
				return // 壊れない
			}
			entry.Durability--
			if entry.Durability <= 0 {
				bus.Publish(MsgLog{Text: def.name + "は壊れてしまった！"})
				a.Inventory = append(a.Inventory[:i], a.Inventory[i+1:]...)
			}
			return
		}
	}
}

func (a *Actor) GetCurrentHP() int {
	return a.HP
}

func (a *Actor) GetCurrentMP() int {
	return a.MP
}

func (a *Actor) GetRewardXP() int {
	return 0
}

func (a *Actor) OnDeath(bus *EventBus) {
	bus.Publish(MsgDropChest{X: a.X, Y: a.Y, Inventory: a.Inventory})
}

func (a *Actor) GainXP(bus *EventBus, amount int) {
	bus.Publish(MsgXP{Actor: a, Amount: amount})
}

type UserPlayer struct {
	Actor
	Name string
}

func (a *UserPlayer) GetName() string {
	return a.Name
}

func NewUserPlayer(name string, race Race, level int, str, wis, fai, vit, agi, luk int) *UserPlayer {
	nextXP := int(10 * math.Pow(float64(level), 1.5))
	hp := playerMaxHP + vit*2
	mp := 5 + wis*5
	return &UserPlayer{
		Actor: Actor{
			ID:        1,
			HP:        hp,
			MaxHP:     hp,
			MP:        mp,
			MaxMP:     mp,
			Dir:       DirDown,
			Element:   ElementNone,
			Race:      race,
			Level:     level,
			XP:        0,
			XPToNext:  nextXP,
			Str:       str,
			Wis:       wis,
			Fai:       fai,
			Vit:       vit,
			Agi:       agi,
			Luk:       luk,
			Relations: make(map[int64]int),
		},
		Name: name,
	}
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
	Player         *UserPlayer
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
	Bus            *EventBus

	menuOpen   bool
	menuCursor int
	menuItems  []actionItem

	mapItems []MapItem

	frame           int     // 毎フレームインクリメント（アニメ用）
	AnimSpeed       float64 // アニメーション速度
	animAccumulator float64 // 速度調整用蓄積

	confirmQuit bool // タイトルへ戻る確認ダイアログ表示中

	nextScene Scene // メッセージで要求された次のシーン
}

func (g *GameScene) GetMap() *[mapWidth][mapHeight]TileType { return &g.worldMap }
func (g *GameScene) IsExplored(x, y int) bool               { return g.explored[x][y] }
func (g *GameScene) SetExplored(x, y int, v bool)           { g.explored[x][y] = v }
func (g *GameScene) GetFloor() int                          { return g.floor }
func (g *GameScene) GetSeed() int64                         { return g.mapSeed }
func (g *GameScene) GetUnitAt(x, y int) Battler             { return g.unitAt(x, y) }
func (g *GameScene) GetPlayState() PlayState                { return g.playState }
func (g *GameScene) SetPlayState(s PlayState)               { g.playState = s }
func (g *GameScene) AddProjectile(p Projectile)             { g.projectiles = append(g.projectiles, p) }

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

// IDをもとに敵を削除する
func (g *GameScene) RemoveEnemy(ID int64) {
	for idx, e := range g.enemies {
		if e.ID == ID {
			g.enemies = append(g.enemies[:idx], g.enemies[idx+1:]...)
		}
	}

}

func NewGameScene() *GameScene {
	p := NewUserPlayer("あなた", RaceHuman, 1, 1, 1, 1, 1, 1, 1)
	return newGameSceneWithState(p, 0, 0, false, nil)
}

func newGameSceneWithState(player *UserPlayer, turnCount, floor int, fromBelow bool, log []string) *GameScene {
	g := &GameScene{
		Player:     player,
		turnCount:  turnCount,
		floor:      floor,
		messageLog: log,
		Combat:     &CombatManager{},
		AnimSpeed:  1.0,
	}
	g.Bus = NewEventBus()
	g.Bus.OnLog(func(msg string) {
		g.message = msg
		g.messageLog = append(g.messageLog, msg)
		if len(g.messageLog) > 1000 {
			g.messageLog = g.messageLog[len(g.messageLog)-1000:]
		}
	})
	g.Bus.OnSFX(func(pcm []byte) {
		playSFX(pcm)
	})
	g.Bus.OnDamage(func(msg MsgDamage) {
		// 必要に応じて演出などを追加
	})
	g.Bus.OnDeath(func(msg MsgDeath) {
		if msg.Battler.GetID() == g.Player.ID {
			g.playState = StateDead
			g.pushMessage("あなたは力尽きた...")
		} else {
			g.RemoveEnemy(msg.Battler.GetID())
			g.pushMessage(fmt.Sprintf("%sは死亡した", msg.Battler.GetName()))
		}
	})
	g.Bus.OnXP(func(msg MsgXP) {
		g.actorGainXP(msg.Actor, msg.Amount)
	})
	g.Bus.OnChest(func(msg MsgDropChest) {
		g.dropChest(msg.X, msg.Y, msg.Inventory)
	})
	g.Bus.OnChangeFloor(func(msg MsgChangeFloor) {
		if msg.Direction > 0 {
			g.nextScene = newGameSceneWithState(g.Player, g.turnCount, msg.CurrentFloor+1, false, g.messageLog)
		} else {
			g.nextScene = newGameSceneWithState(g.Player, g.turnCount, msg.CurrentFloor-1, true, g.messageLog)
		}
	})
	g.Bus.OnTransition(func(msg MsgTransition) {
		g.nextScene = msg.Next
	})

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
			a.Str = min(a.Str, MaxStr)
		case 1:
			a.Wis++
			a.Wis = min(a.Wis, MaxWis)
			a.MP += 5
			a.MaxMP += 5
		case 2:
			a.Fai++
			a.Fai = min(a.Fai, MaxFai)
		case 3:
			a.Vit++
			a.Vit = min(a.Vit, MaxVit)
			a.HP += 2
			a.MaxHP += 2
		case 4:
			a.Agi++
			a.Agi = min(a.Agi, MaxAgi)
		case 5:
			a.Luk++
			a.Luk = min(a.Luk, MaxLuk)
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
	g.Bus.Publish(MsgLog{Text: msg})
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
	if g.nextScene != nil {
		return g.nextScene, nil
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
			g.Bus.Publish(MsgSFX{PCM: sfxStairUpPCM})
		} else {
			g.pushMessage("宝がない！まだ帰れない。")
		}
	case StairsUp:
		g.Bus.Publish(MsgSFX{PCM: sfxStairUpPCM})
		g.Bus.Publish(MsgChangeFloor{CurrentFloor: g.floor, Direction: -1})
	case StairsDown:
		g.Bus.Publish(MsgSFX{PCM: sfxStairDownPCM})
		g.Bus.Publish(MsgChangeFloor{CurrentFloor: g.floor, Direction: 1})
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
