package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 480
	tileSize     = 16
	mapWidth     = 40
	mapHeight    = 25
)

// タイルの種類
type TileType int

const (
	Wall TileType = iota
	Floor
	Stairs
	Treasure
)

// ゲームの状態
type GameState int

const (
	StatePlaying GameState = iota
	StateWin
)

type Game struct {
	worldMap      [mapWidth][mapHeight]TileType
	explored      [mapWidth][mapHeight]bool
	playerX       int
	playerY       int
	hasTreasure   bool
	gameState     GameState
	message       string
	turnCount     int
}

func NewGame() *Game {
	g := &Game{}
	g.generateMap()
	g.message = "ダンジョンを探索し、宝を見つけて戻ってこい！"
	return g
}

// ダンジョン生成（シンプルな部屋生成アルゴリズム）
func (g *Game) generateMap() {
	rand.Seed(time.Now().UnixNano())

	// 全て壁で埋める
	for x := 0; x < mapWidth; x++ {
		for y := 0; y < mapHeight; y++ {
			g.worldMap[x][y] = Wall
		}
	}

	// 部屋をいくつか作成
	rooms := 0
	for i := 0; i < 15; i++ {
		w := rand.Intn(6) + 4
		h := rand.Intn(6) + 4
		x := rand.Intn(mapWidth - w - 1) + 1
		y := rand.Intn(mapHeight - h - 1) + 1

		// 部屋を彫る
		for rx := x; rx < x+w; rx++ {
			for ry := y; ry < y+h; ry++ {
				g.worldMap[rx][ry] = Floor
			}
		}

		// 最初の部屋にプレイヤーを配置
		if rooms == 0 {
			g.playerX = x + w/2
			g.playerY = y + h/2
			// 入り口（階段）を配置
			g.worldMap[g.playerX][g.playerY] = Stairs
		}
		rooms++
	}

	// 最後にどこかに宝を置く
	for {
		tx := rand.Intn(mapWidth)
		ty := rand.Intn(mapHeight)
		if g.worldMap[tx][ty] == Floor && (tx != g.playerX || ty != g.playerY) {
			g.worldMap[tx][ty] = Treasure
			break
		}
	}
}

func (g *Game) Update() error {
	if g.gameState == StateWin {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			*g = *NewGame()
		}
		return nil
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
				g.playerX = newX
				g.playerY = newY
				g.turnCount++
				g.updateExplored()
				g.checkTile()
			}
		}
	}

	return nil
}

func (g *Game) updateExplored() {
	// プレイヤーの周囲を探索済みにする（簡易FOV）
	radius := 4
	for x := g.playerX - radius; x <= g.playerX+radius; x++ {
		for y := g.playerY - radius; y <= g.playerY+radius; y++ {
			if x >= 0 && x < mapWidth && y >= 0 && y < mapHeight {
				g.explored[x][y] = true
			}
		}
	}
}

func (g *Game) checkTile() {
	tile := g.worldMap[g.playerX][g.playerY]
	switch tile {
	case Treasure:
		g.hasTreasure = true
		g.worldMap[g.playerX][g.playerY] = Floor
		g.message = "宝を手に入れた！入り口の階段まで戻れ！"
	case Stairs:
		if g.hasTreasure {
			g.gameState = StateWin
			g.message = fmt.Sprintf("脱出成功！ターン数: %d (Rキーで再挑戦)", g.turnCount)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 背景描画
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
			}

			// 描画位置計算
			rect := ebiten.NewImage(tileSize-1, tileSize-1)
			rect.Fill(clr)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))
			screen.DrawImage(rect, op)
		}
	}

	// プレイヤー描画
	playerClr := color.RGBA{255, 100, 100, 255}
	if g.hasTreasure {
		playerClr = color.RGBA{255, 255, 100, 255} // 宝を持っている時は色を変える
	}
	pRect := ebiten.NewImage(tileSize-1, tileSize-1)
	pRect.Fill(playerClr)
	pOp := &ebiten.DrawImageOptions{}
	pOp.GeoM.Translate(float64(g.playerX*tileSize), float64(g.playerY*tileSize))
	screen.DrawImage(pRect, pOp)

	// UIメッセージ
	ebitenutil.DebugPrint(screen, fmt.Sprintf("\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n  %s", g.message))
	ebitenutil.DebugPrint(screen, fmt.Sprintf("  Turn: %d", g.turnCount))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	game := NewGame()
	game.updateExplored()
	ebiten.SetWindowTitle("Ebiten Roguelike")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
