package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/animation"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/event"
	"github.com/gentaman/myrogue/internal/core/world"
)

func (g *GameScene) processEvents(events []event.Event) {
	for _, ev := range events {
		switch e := ev.(type) {
		case event.EventLog:
			g.pushMessage(e.Text)
		case event.EventSFX:
			if g.Audio != nil {
				g.Audio.PlaySFX(e.ID)
			}
		case event.EventDeath:
			if e.Entity == g.Player {
				g.PlayState = StateDead
				g.pushMessage("あなたは力尽きた...")
			} else {
				faction, _ := g.Factions.Get(e.Entity)
				if faction != nil && faction.Faction == component.FactionAlly {
					g.pushMessage(fmt.Sprintf("%sは倒れた...", e.Name))
				} else {
					g.pushMessage(fmt.Sprintf("%sは死亡した", e.Name))
				}
				cascaded := g.Resolver.ProcessDeath(e)
				g.processEvents(cascaded)
			}
		case event.EventXP:
			cascaded := g.Resolver.ProcessXP(e)
			if e.Entity == g.Player {
				g.pushMessage(fmt.Sprintf("%d の経験値を獲得！", e.Amount))
			}
			g.processEvents(cascaded)
		case event.EventLevelUp:
			if e.Entity == g.Player {
				g.pushMessage(fmt.Sprintf("レベル %d に上がった！", e.NewLevel))
			}
		case event.EventAttack:
			g.Resolver.ProcessAttack(e)
			// アタッカーの向きを更新
			if pos, ok := g.Positions.Get(e.Attacker); ok {
				pos.Dir = e.Dir
			}
			// アタッカーの攻撃アニメーション
			if anim, ok := g.Anims.Get(e.Attacker); ok {
				anim.AttackAnim = component.AttackAnimFrames
			}
			// ディフェンダーの被ダメージアニメーション
			if anim, ok := g.Anims.Get(e.Defender); ok {
				anim.DamageAnim = component.DamageAnimFrames
			}
		case event.EventMoved:
			if pos, ok := g.Positions.Get(e.Entity); ok {
				pos.X, pos.Y = e.ToX, e.ToY
				pos.Dir = e.Dir
			}
		case event.EventDirChanged:
			if pos, ok := g.Positions.Get(e.Entity); ok {
				pos.Dir = e.Dir
			}
		case event.EventSwap:
			p1, ok1 := g.Positions.Get(e.Entity1)
			p2, ok2 := g.Positions.Get(e.Entity2)
			if ok1 && ok2 {
				p1.X, p1.Y, p2.X, p2.Y = p2.X, p2.Y, p1.X, p1.Y
			}
		case event.EventWait:
			// 待機時の処理があればここに書く
		case event.EventDrop:
			g.World.Items = append(g.World.Items, world.MapItem{X: e.X, Y: e.Y, Inventory: e.Items})
		case event.EventEquip:
			g.Resolver.ProcessEquip(e)
		case event.EventItemConsume:
			g.Resolver.ProcessItemConsume(e)
		case event.EventProjectile:
			g.AnimQueue.Add(animation.Projectile{
				StartX:      e.StartX * TileSize,
				StartY:      e.StartY * TileSize,
				EndX:        e.EndX * TileSize,
				EndY:        e.EndY * TileSize,
				TotalFrames: e.TotalFrames,
				ColorHex:    e.ColorHex,
				IsFlash:     e.IsFlash,
			})
		case event.EventFloorChange:
			g.doChangeFloor(e.Direction)
		case event.EventMP:
			g.Resolver.ProcessMP(e)
		case event.EventHeal:
			g.Resolver.ProcessHeal(e)
		case event.EventStatusEffect:
			g.Resolver.ProcessStatusEffect(e)
		case event.EventWin:
			g.PlayState = StateWin
			score := g.calcScore() // Use existing complex score calculation
			g.pushMessage(fmt.Sprintf("脱出成功！ スコア: %d", score))
			if g.Audio != nil {
				g.Audio.PlaySFX("stair_up")
			}
		case event.EventRevealMap:
			for x := 0; x < world.MapWidth; x++ {
				for y := 0; y < world.MapHeight; y++ {
					g.World.Explored[x][y] = true
				}
			}
		}
	}
}
