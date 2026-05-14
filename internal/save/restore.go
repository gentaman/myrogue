package save

import (
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

type GameRestorer interface {
	RestoreInit(playerID entity.ID, floor, turnCount int, mapSeed int64, messageLog []string)
	RestoreMap(explored []ExploredTile, items []MapItemSnap)
	RestoreEntity(es EntitySnapshot)
}

func Restore(snap *Snapshot, r GameRestorer) {
	r.RestoreInit(entity.ID(snap.PlayerID), snap.Floor, snap.TurnCount, snap.MapSeed, snap.MessageLog)
	r.RestoreMap(snap.Explored, snap.MapItems)
	for _, es := range snap.Entities {
		r.RestoreEntity(es)
	}
}

func RestoreEntityComponents(
	es EntitySnapshot,
	id entity.ID,
	positions *entity.Store[component.Position],
	stats *entity.Store[component.Stats],
	factions *entity.Store[component.FactionComp],
	ais *entity.Store[component.AI],
	inventories *entity.Store[component.Inventory],
	anims *entity.Store[component.AnimState],
	appearances *entity.Store[component.Appearance],
	names *entity.Store[component.Name],
	rewards *entity.Store[component.Reward],
	elements *entity.Store[component.Element],
	races *entity.Store[component.Race],
) {
	if es.Position != nil {
		positions.Set(id, &component.Position{
			X: es.Position.X, Y: es.Position.Y,
			Dir: component.Dir(es.Position.Dir),
		})
	}

	if es.Stats != nil {
		stats.Set(id, &component.Stats{
			HP: es.Stats.HP, MaxHP: es.Stats.MaxHP,
			MP: es.Stats.MP, MaxMP: es.Stats.MaxMP,
			Level: es.Stats.Level, XP: es.Stats.XP, XPToNext: es.Stats.XPToNext,
			Str: es.Stats.Str, Wis: es.Stats.Wis, Fai: es.Stats.Fai,
			Vit: es.Stats.Vit, Agi: es.Stats.Agi, Luk: es.Stats.Luk,
		})
	}

	if es.Faction != "" {
		factions.Set(id, component.NewFactionComp(component.FactionFromString(es.Faction)))
	}

	if es.AI != nil {
		ais.Set(id, &component.AI{
			Personality:       component.Personality(es.AI.Personality),
			Order:             component.CompanionOrder(es.AI.Order),
			State:             component.AlertState(es.AI.State),
			FriendlyFire:      es.AI.FriendlyFire,
			NeutralThreshold:  es.AI.NeutralThreshold,
			FriendlyThreshold: es.AI.FriendlyThreshold,
		})
	}

	if len(es.Inventory) > 0 {
		inventories.Set(id, &component.Inventory{Items: es.Inventory})
	} else {
		inventories.Set(id, &component.Inventory{})
	}

	anims.Set(id, &component.AnimState{})
	names.Set(id, &component.Name{Value: es.Name})
	rewards.Set(id, &component.Reward{XP: es.RewardXP})

	el := component.ElementFromString(es.Element)
	elements.Set(id, &el)

	race := component.RaceFromString(es.Race)
	races.Set(id, &race)

	appearances.Set(id, &component.Appearance{
		DefID:     es.DefID,
		ColorHex:  es.ColorHex,
		HasSprite: es.HasSprite,
	})
}

func RestoreMapExplored(gm *world.GameMap, explored []ExploredTile) {
	for _, t := range explored {
		if gm.InBounds(t.X, t.Y) {
			gm.Explored[t.X][t.Y] = true
		}
	}
}

func RestoreMapItems(gm *world.GameMap, items []MapItemSnap) {
	gm.Items = make([]world.MapItem, len(items))
	for i, it := range items {
		gm.Items[i] = world.MapItem{X: it.X, Y: it.Y, Inventory: it.Items}
	}
}
