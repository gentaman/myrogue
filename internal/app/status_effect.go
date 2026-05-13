package app

import (
	"fmt"

	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
)

func (g *GameScene) tickStatusEffects() {
	g.StatusEffects.Each(func(id entity.ID, se *component.StatusEffects) {
		if !g.Entities.IsAlive(id) {
			return
		}
		stats := g.GetStats(id)
		if stats == nil {
			return
		}
		name := g.GetName(id)

		for _, eff := range se.Effects {
			switch eff.Type {
			case component.StatusPoison:
				dmg := stats.MaxHP / 10
				if dmg < 1 {
					dmg = 1
				}
				stats.HP -= dmg
				if id == g.Player {
					g.pushMessage(fmt.Sprintf("毒で %d ダメージを受けた！", dmg))
				}
				if stats.HP <= 0 {
					g.pushMessage(fmt.Sprintf("%sは毒で力尽きた...", name))
					g.Entities.Destroy(id)
					if id == g.Player {
						g.PlayState = StateDead
					}
				}
			case component.StatusRegen:
				heal := stats.MaxHP / 10
				if heal < 1 {
					heal = 1
				}
				stats.HP += heal
				if stats.HP > stats.MaxHP {
					stats.HP = stats.MaxHP
				}
				if id == g.Player {
					g.pushMessage(fmt.Sprintf("再生で %d 回復した。", heal))
				}
			}
		}

		expired := se.Tick()
		for _, t := range expired {
			if id == g.Player {
				g.pushMessage(fmt.Sprintf("%sが治った。", component.StatusNames[t]))
			}
		}
	})
}

func (g *GameScene) ApplyStatus(target entity.ID, status component.StatusType, duration int) {
	se, ok := g.StatusEffects.Get(target)
	if !ok {
		se = &component.StatusEffects{}
		g.StatusEffects.Set(target, se)
	}
	se.Add(status, duration)
	name := g.GetName(target)
	g.pushMessage(fmt.Sprintf("%sは%s状態になった！", name, component.StatusNames[status]))
}

func (g *GameScene) isParalyzed(id entity.ID) bool {
	se, ok := g.StatusEffects.Get(id)
	if !ok {
		return false
	}
	return se.Has(component.StatusParalyze)
}

func (g *GameScene) isConfused(id entity.ID) bool {
	se, ok := g.StatusEffects.Get(id)
	if !ok {
		return false
	}
	return se.Has(component.StatusConfuse)
}

func (g *GameScene) isSleeping(id entity.ID) bool {
	se, ok := g.StatusEffects.Get(id)
	if !ok {
		return false
	}
	return se.Has(component.StatusSleep)
}
