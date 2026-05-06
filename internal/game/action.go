package game

const (
	menuKindAttack = iota
	menuKindStairs
	menuKindItem
	menuKindClose
)

type actionItem struct {
	label   string
	enabled bool
	kind    int
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

	var attackLabel string
	if hasAdj {
		attackLabel = "攻撃する"
	} else {
		attackLabel = "攻撃する（近くに敵なし）"
	}

	var stairLabel string
	switch stair {
	case Stairs:
		if g.hasTreasure {
			stairLabel = "上り階段を使う（脱出）"
		} else {
			stairLabel = "上り階段を使う（宝がない）"
		}
	case StairsUp:
		stairLabel = "上り階段を使う"
	case StairsDown:
		stairLabel = "下り階段を使う"
	default:
		stairLabel = "階段を使う（階段の上にいない）"
	}

	g.menuItems = []actionItem{
		{label: attackLabel, enabled: hasAdj, kind: menuKindAttack},
		{label: stairLabel, enabled: stair != Wall && (stair != Stairs || g.hasTreasure), kind: menuKindStairs},
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
		g.attackEnemy(ex, ey)
		g.moveEnemies()
		if g.playState == StateDead {
			return nil, nil
		}
		g.trySpawnEnemyPerTurn()
	case menuKindStairs:
		g.turnCount++
		if next, err := g.checkTile(); next != nil || err != nil {
			return next, err
		}
		g.moveEnemies()
		if g.playState == StateDead {
			return nil, nil
		}
		g.trySpawnEnemyPerTurn()
	}
	return nil, nil
}
