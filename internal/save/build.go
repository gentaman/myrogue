package save

import (
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
)

type GameAccess interface {
	GetPlayerID() entity.ID
	GetFloor() int
	GetTurnCount() int
	GetMapSeed() int64
	GetMessageLog() []string
	GetGameMap() *world.GameMap
	AllEntities() []entity.ID
	GetPosition(id entity.ID) *component.Position
	GetStats(id entity.ID) *component.Stats
	GetFaction(id entity.ID) *component.FactionComp
	GetAI(id entity.ID) *component.AI
	GetInventory(id entity.ID) *component.Inventory
	GetElement(id entity.ID) component.Element
	GetRace(id entity.ID) component.Race
	GetAppearance(id entity.ID) *component.Appearance
	GetName(id entity.ID) string
	GetRewardXP(id entity.ID) int
	IsAlive(id entity.ID) bool
}

func Build(g GameAccess) *Snapshot {
	snap := &Snapshot{
		SchemaVersion: SchemaVersion,
		Floor:         g.GetFloor(),
		TurnCount:     g.GetTurnCount(),
		MapSeed:       g.GetMapSeed(),
		PlayerID:      int64(g.GetPlayerID()),
		MessageLog:    g.GetMessageLog(),
	}

	gm := g.GetGameMap()

	for _, id := range g.AllEntities() {
		if !g.IsAlive(id) {
			continue
		}
		es := buildEntity(g, id)
		snap.Entities = append(snap.Entities, es)
	}

	for _, it := range gm.Items {
		if len(it.Inventory) > 0 {
			snap.MapItems = append(snap.MapItems, MapItemSnap{
				X: it.X, Y: it.Y, Items: it.Inventory,
			})
		}
	}

	for x := 0; x < world.MapWidth; x++ {
		for y := 0; y < world.MapHeight; y++ {
			if gm.Explored[x][y] {
				snap.Explored = append(snap.Explored, ExploredTile{X: x, Y: y})
			}
		}
	}

	return snap
}

func buildEntity(g GameAccess, id entity.ID) EntitySnapshot {
	es := EntitySnapshot{
		ID:   int64(id),
		Name: g.GetName(id),
	}

	fc := g.GetFaction(id)
	if fc != nil {
		es.Faction = fc.Faction.String()
		switch fc.Faction {
		case component.FactionPlayer:
			es.Kind = "player"
		case component.FactionAlly:
			es.Kind = "companion"
		case component.FactionEnemy:
			es.Kind = "enemy"
		}
	}

	pos := g.GetPosition(id)
	if pos != nil {
		es.Position = &PositionSnap{X: pos.X, Y: pos.Y, Dir: int(pos.Dir)}
	}

	stats := g.GetStats(id)
	if stats != nil {
		es.Stats = &StatsSnap{
			HP: stats.HP, MaxHP: stats.MaxHP,
			MP: stats.MP, MaxMP: stats.MaxMP,
			Level: stats.Level, XP: stats.XP, XPToNext: stats.XPToNext,
			Str: stats.Str, Wis: stats.Wis, Fai: stats.Fai,
			Vit: stats.Vit, Agi: stats.Agi, Luk: stats.Luk,
		}
	}

	aiComp := g.GetAI(id)
	if aiComp != nil {
		es.AI = &AISnap{
			Personality:       int(aiComp.Personality),
			Order:             int(aiComp.Order),
			State:             int(aiComp.State),
			FriendlyFire:      aiComp.FriendlyFire,
			NeutralThreshold:  aiComp.NeutralThreshold,
			FriendlyThreshold: aiComp.FriendlyThreshold,
		}
	}

	inv := g.GetInventory(id)
	if inv != nil && len(inv.Items) > 0 {
		es.Inventory = inv.Items
	}

	el := g.GetElement(id)
	es.Element = el.String()

	race := g.GetRace(id)
	es.Race = race.String()

	app := g.GetAppearance(id)
	if app != nil {
		es.DefID = app.DefID
		es.ColorHex = app.ColorHex
		es.HasSprite = app.HasSprite
	}

	es.RewardXP = g.GetRewardXP(id)

	return es
}
