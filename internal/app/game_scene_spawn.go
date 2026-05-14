package app

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/world"
	"github.com/gentaman/myrogue/internal/mapgen"
)

func (g *GameScene) spawnPlayer(reg *content.Registry) {
	def := reg.Players[0]
	hp := def.HP + def.Vit*2
	mp := def.MP + def.Wis*5
	g.Player = entity.ID(1)
	g.Entities.CreateWithID(g.Player)

	g.Positions.Set(g.Player, &component.Position{Dir: component.DirDown})
	g.Stats.Set(g.Player, &component.Stats{
		HP: hp, MaxHP: hp,
		MP: mp, MaxMP: mp,
		Level: 1, XPToNext: 10,
		Str: def.Str, Wis: def.Wis, Fai: def.Fai,
		Vit: def.Vit, Agi: def.Agi, Luk: def.Luk,
	})
	g.Factions.Set(g.Player, component.NewFactionComp(component.FactionPlayer))
	g.Anims.Set(g.Player, &component.AnimState{})
	g.Names.Set(g.Player, &component.Name{Value: "あなた"})
	g.Inventories.Set(g.Player, &component.Inventory{MaxCarryWeight: reg.MaxCarryWeight})
	el := def.Element
	g.Elements.Set(g.Player, &el)
	race := def.Race
	g.Races.Set(g.Player, &race)
	g.Appearances.Set(g.Player, &component.Appearance{DefID: "player", HasSprite: true})
	g.Skills.Set(g.Player, &component.SkillComp{Skills: def.Skills})
}

func (g *GameScene) generateFloor(floor int) {
	floorDef := g.registry.GetFloorDef(floor)
	seed := time.Now().UnixNano()
	g.World = mapgen.Generate(floor, floorDef, g.registry, seed, g)

	if len(g.World.Rooms) > 0 {
		r := g.World.Rooms[0]
		pos := g.playerPos()
		pos.X = r.CenterX()
		pos.Y = r.CenterY()
	}

	g.spawnInitialEnemies()
	if g.registry.SelectedCompanion != "" {
		g.spawnCompanion(g.registry.SelectedCompanion)
	}
	world.UpdateVisibility(g.World, g.playerPos().X, g.playerPos().Y)
}

func (g *GameScene) spawnInitialEnemies() {
	floorDef := g.registry.GetFloorDef(g.World.Floor)
	playerRoom := g.World.RoomOf(g.playerPos().X, g.playerPos().Y)
	count := len(g.World.Rooms) / 2
	for i := 0; i < count; i++ {
		g.trySpawnEnemy(floorDef, playerRoom)
	}
}

func (g *GameScene) trySpawnEnemy(floorDef *content.FloorDef, playerRoom int) {
	if len(floorDef.EnemyPool) == 0 {
		return
	}
	weights := make([]int, len(floorDef.EnemyPool))
	for i, entry := range floorDef.EnemyPool {
		weights[i] = entry.Weight
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for attempt := 0; attempt < 50; attempt++ {
		roomIdx := rng.Intn(len(g.World.Rooms))
		if roomIdx == playerRoom {
			continue
		}
		r := g.World.Rooms[roomIdx]
		ex, ey := r.X+rng.Intn(r.W), r.Y+rng.Intn(r.H)
		if g.World.Tiles[ex][ey] != world.Floor || g.UnitAt(ex, ey) != entity.InvalidID {
			continue
		}
		poolIdx := content.WeightedChoice(rng, weights)
		if poolIdx < 0 {
			return
		}
		defID := floorDef.EnemyPool[poolIdx].ID
		def, ok := g.registry.GetEnemyDef(defID)
		if !ok {
			return
		}
		g.spawnEnemyAt(def, ex, ey)
		return
	}
}

func (g *GameScene) spawnEnemyAt(def *content.ActorDef, x, y int) entity.ID {
	id := g.Entities.Create()
	g.Positions.Set(id, &component.Position{X: x, Y: y, Dir: component.DirDown})
	g.Stats.Set(id, &component.Stats{
		HP: def.HP, MaxHP: def.HP,
		MP: def.MP, MaxMP: def.MP,
		Level:    1 + g.World.Floor*2,
		XPToNext: 10,
		Str:      def.Str, Wis: def.Wis, Fai: def.Fai,
		Vit: def.Vit, Agi: def.Agi, Luk: def.Luk,
	})
	g.Factions.Set(id, component.NewFactionComp(component.FactionEnemy))
	g.AIs.Set(id, &component.AI{
		Personality:       def.Personality,
		NeutralThreshold:  def.NeutralThreshold,
		FriendlyThreshold: def.FriendlyThreshold,
	})
	g.Anims.Set(id, &component.AnimState{})
	g.Names.Set(id, &component.Name{Value: def.Name})
	g.Rewards.Set(id, &component.Reward{XP: def.XP})
	el := def.Element
	g.Elements.Set(id, &el)
	race := def.Race
	g.Races.Set(id, &race)
	g.Appearances.Set(id, &component.Appearance{DefID: def.ID, ColorHex: def.ColorHex})
	g.Inventories.Set(id, &component.Inventory{})
	g.Skills.Set(id, &component.SkillComp{Skills: def.Skills})
	return id
}

func (g *GameScene) spawnCompanion(defID string) {
	def, ok := g.registry.GetCompanionDef(defID)
	if !ok {
		return
	}
	pPos := g.playerPos()
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for _, d := range dirs {
		nx, ny := pPos.X+d[0], pPos.Y+d[1]
		if g.World.InBounds(nx, ny) && g.World.IsPassable(nx, ny) && g.UnitAt(nx, ny) == entity.InvalidID {
			id := g.Entities.Create()
			g.Positions.Set(id, &component.Position{X: nx, Y: ny, Dir: component.DirDown})
			g.Stats.Set(id, &component.Stats{
				HP: def.HP, MaxHP: def.HP,
				MP: def.MP, MaxMP: def.MP,
				Level: 1, XPToNext: 10,
				Str: def.Str, Wis: def.Wis, Fai: def.Fai,
				Vit: def.Vit, Agi: def.Agi, Luk: def.Luk,
			})
			g.Factions.Set(id, component.NewFactionComp(component.FactionAlly))
			g.AIs.Set(id, &component.AI{
				Personality:       def.Personality,
				Order:             component.OrderFollow,
				FriendlyFire:      def.FriendlyFire,
				NeutralThreshold:  def.NeutralThreshold,
				FriendlyThreshold: def.FriendlyThreshold,
			})
			g.Anims.Set(id, &component.AnimState{})
			g.Names.Set(id, &component.Name{Value: def.Name})
			g.Rewards.Set(id, &component.Reward{XP: 0})
			el := def.Element
			g.Elements.Set(id, &el)
			race := def.Race
			g.Races.Set(id, &race)
			g.Appearances.Set(id, &component.Appearance{DefID: def.ID, ColorHex: def.ColorHex})
			g.Inventories.Set(id, &component.Inventory{})
			g.Skills.Set(id, &component.SkillComp{Skills: def.Skills})
			return
		}
	}
}

func (g *GameScene) tryChangeFloor(direction int) {
	pPos := g.playerPos()
	var leftBehind []string
	allies := g.EntitiesWithFaction(component.FactionAlly)
	for _, id := range allies {
		aPos := g.GetPosition(id)
		if aPos == nil {
			continue
		}
		dist := abs(aPos.X-pPos.X) + abs(aPos.Y-pPos.Y)
		if dist > 3 {
			leftBehind = append(leftBehind, g.GetName(id))
		}
	}

	if len(leftBehind) > 0 {
		g.ConfirmStair = true
		g.StairLeft = leftBehind
		g.PendingDir = direction
	} else {
		act := &action.FloorChangeAction{Direction: direction}
		g.processEvents(act.Execute(g.Player, g))
	}
}

func (g *GameScene) doChangeFloor(direction int) {
	nextFloor := g.World.Floor + direction
	if nextFloor < 0 {
		nextFloor = 0
	}

	// Carry over companions that are close enough
	pPos := g.playerPos()
	allies := g.EntitiesWithFaction(component.FactionAlly)
	var companionDefs []string
	for _, id := range allies {
		aPos := g.GetPosition(id)
		if aPos == nil {
			continue
		}
		dist := abs(aPos.X-pPos.X) + abs(aPos.Y-pPos.Y)
		if dist <= 3 {
			a, ok := g.Appearances.Get(id)
			if ok {
				companionDefs = append(companionDefs, a.DefID)
			}
		}
	}

	// Generate new floor
	floorDef := g.registry.GetFloorDef(nextFloor)
	seed := time.Now().UnixNano()
	g.World = mapgen.Generate(nextFloor, floorDef, g.registry, seed, g)

	// Remove old non-player entities
	g.Positions.Each(func(id entity.ID, pos *component.Position) {
		if id != g.Player {
			g.Entities.Destroy(id)
		}
	})

	// Place player
	if direction > 0 {
		// Going down — place at first room
		if len(g.World.Rooms) > 0 {
			r := g.World.Rooms[0]
			pPos.X = r.CenterX()
			pPos.Y = r.CenterY()
		}
	} else {
		// Going up — place at StairsDown if exists
		found := false
		for x := 0; x < world.MapWidth && !found; x++ {
			for y := 0; y < world.MapHeight && !found; y++ {
				if g.World.Tiles[x][y] == world.StairsDown {
					pPos.X, pPos.Y = x, y
					found = true
				}
			}
		}
		if !found && len(g.World.Rooms) > 0 {
			r := g.World.Rooms[0]
			pPos.X = r.CenterX()
			pPos.Y = r.CenterY()
		}
	}

	// Spawn enemies and companions
	g.spawnInitialEnemies()
	for _, defID := range companionDefs {
		g.spawnCompanion(defID)
	}

	world.UpdateVisibility(g.World, pPos.X, pPos.Y)
	g.pushMessage(fmt.Sprintf("フロア %d に移動した。", nextFloor+1))
}
