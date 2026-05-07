package game

import "fmt"

const (
	menuKindAttack = iota
	menuKindExamine
	menuKindStairs
	menuKindWait
	menuKindItem
	menuKindClose
)

type actionItem struct {
	label   string
	enabled bool
	kind    int
}

// 足元にアイテムがあるかチェック
func (g *GameScene) itemAtFeet() (int, bool) {
	for i, it := range g.mapItems {
		if it.x == g.playerX && it.y == g.playerY {
			return i, true
		}
	}
	return -1, false
}

// プレイヤーの正面にいる敵を返す
func (g *GameScene) adjacentEnemy() (int, int, bool) {
	dx, dy := g.playerDir.delta()
	nx, ny := g.playerX+dx, g.playerY+dy
	if g.isEnemyAt(nx, ny) {
		return nx, ny, true
	}
	return 0, 0, false
}

// 現在地の階段種別を返す（なければ Wall）
func (g *GameScene) currentStairType() TileType {
	t := g.worldMap[g.playerX][g.playerY]
	switch t {
	case Stairs, StairsUp, StairsDown:
		return t
	}
	return Wall
}

// 文脈に応じたアクションメニューを構築する
func (g *GameScene) buildMenu() {
	_, _, hasAdj := g.adjacentEnemy()
	stair := g.currentStairType()
	_, hasItem := g.itemAtFeet()

	var attackLabel string
	if hasAdj {
		attackLabel = "攻撃する"
	} else {
		attackLabel = "攻撃する（正面に敵なし）"
	}

	var examineLabel string
	if hasItem {
		examineLabel = "足元のアイテムを拾う"
	} else if stair != Wall {
		examineLabel = "階段を調べる"
	} else {
		examineLabel = "調べる（何もなし）"
	}

	itemLabel := fmt.Sprintf("アイテムを使う（重量: %d/%d）", g.currentWeight(), maxCarryWeight)
	if len(g.inventory) == 0 {
		itemLabel = "アイテムを使う（所持なし）"
	}

	g.menuItems = []actionItem{
		{label: attackLabel, enabled: hasAdj, kind: menuKindAttack},
		{label: examineLabel, enabled: hasItem || stair != Wall, kind: menuKindExamine},
		{label: itemLabel, enabled: len(g.inventory) > 0, kind: menuKindItem},
		{label: "待機する（1ターン消費）", enabled: true, kind: menuKindWait},
		{label: "閉じる", enabled: true, kind: menuKindClose},
	}
	if g.menuCursor >= len(g.menuItems) {
		g.menuCursor = 0
	}
}

// アクションメニューの選択項目を実行する
func (g *GameScene) execMenuItem() (Scene, error) {
	item := g.menuItems[g.menuCursor]
	if !item.enabled {
		return nil, nil
	}
	g.menuOpen = false
	switch item.kind {
	case menuKindAttack:
		ex, ey, _ := g.adjacentEnemy()
		g.turnCount++
		g.playerAttackAnim = attackAnimFrames
		g.attackEnemy(ex, ey)
		g.moveEnemies()
		if g.playState == StateDead {
			return nil, nil
		}
		g.trySpawnEnemyPerTurn()
	case menuKindExamine:
		g.turnCount++
		// アイテムがあれば拾う、なければ階段判定
		if idx, found := g.itemAtFeet(); found {
			g.pickupItem(idx)
		} else if next, err := g.checkTile(); next != nil || err != nil {
			return next, err
		} else {
			g.pushMessage("特に何もない。")
		}
		g.moveEnemies()
		if g.playState == StateDead {
			return nil, nil
		}
		g.trySpawnEnemyPerTurn()
	case menuKindItem:
		return &InventoryScene{game: g}, nil
	case menuKindWait:
		g.turnCount++
		g.pushMessage("待機した。")
		// MP回復
		if g.MP < g.MaxMP {
			g.MP++
		}
		g.moveEnemies()
		if g.playState == StateDead {
			return nil, nil
		}
		g.trySpawnEnemyPerTurn()
	}
	return nil, nil
}
