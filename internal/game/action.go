package game

// アクションメニューの項目
type menuActionKind int

const (
	menuKindAttack menuActionKind = iota
	menuKindExamine
	menuKindItem
	menuKindWait
)

type actionItem struct {
	kind    menuActionKind
	label   string
	enabled bool
}

// 足元にアイテムがあるか判定
func (g *GameScene) itemAtFeet() (int, bool) {
	for i, it := range g.mapItems {
		if it.x == g.Player.X && it.y == g.Player.Y {
			return i, true
		}
	}
	return -1, false
}

// 隣接する敵がいるか判定
func (g *GameScene) adjacentEnemy() (int, int, bool) {
	dx, dy := g.Player.Dir.delta()
	nx, ny := g.Player.X+dx, g.Player.Y+dy
	if g.isEnemyAt(nx, ny) {
		return nx, ny, true
	}
	return -1, -1, false
}

// 足元の階段の種類を返す
func (g *GameScene) currentStairType() TileType {
	t := g.worldMap[g.Player.X][g.Player.Y]
	if t == Stairs || t == StairsUp || t == StairsDown {
		return t
	}
	return Wall // 便宜上 Wall を返す（階段ではない）
}

// アクションメニューの選択項目を構築する
func (g *GameScene) buildMenu() {
	g.menuItems = []actionItem{}

	// 攻撃
	_, _, hasAdj := g.adjacentEnemy()
	g.menuItems = append(g.menuItems, actionItem{
		kind:    menuKindAttack,
		label:   "正面を攻撃",
		enabled: hasAdj,
	})

	// 調べる（拾う・階段）
	label := "足元を調べる"
	if _, found := g.itemAtFeet(); found {
		label = "アイテムを拾う"
	} else {
		switch g.currentStairType() {
		case Stairs, StairsUp:
			label = "上の階へ戻る"
		case StairsDown:
			label = "下の階へ進む"
		}
	}
	g.menuItems = append(g.menuItems, actionItem{
		kind:    menuKindExamine,
		label:   label,
		enabled: true,
	})

	// アイテム
	g.menuItems = append(g.menuItems, actionItem{
		kind:    menuKindItem,
		label:   "道具を使う",
		enabled: len(g.inventory) > 0,
	})

	// 待機
	g.menuItems = append(g.menuItems, actionItem{
		kind:    menuKindWait,
		label:   "その場で待機",
		enabled: true,
	})
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
		g.Player.AttackAnim = attackAnimFrames
		g.attackEnemy(ex, ey)
	case menuKindExamine:
		g.turnCount++
		if idx, found := g.itemAtFeet(); found {
			g.pickupItem(idx)
		} else if next, err := g.checkTile(); next != nil || err != nil {
			return next, err
		} else {
			g.pushMessage("特に何もない。")
		}
	case menuKindItem:
		return &InventoryScene{game: g}, nil
	case menuKindWait:
		g.turnCount++
		g.pushMessage("待機した。")
	}
	return nil, nil
}
