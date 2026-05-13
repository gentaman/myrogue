package app

import (
	"github.com/gentaman/myrogue/internal/animation"
	"github.com/gentaman/myrogue/internal/clock"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/rules"
	"github.com/gentaman/myrogue/internal/core/turn"
	"github.com/gentaman/myrogue/internal/core/world"
	"github.com/gentaman/myrogue/internal/debug"
	"github.com/gentaman/myrogue/internal/mapgen"
	"github.com/gentaman/myrogue/internal/save"
)

type PlayState int

const (
	StatePlaying PlayState = iota
	StateWin
	StateDead
)

type GameScene struct {
	Entities      *entity.Manager
	Positions     *entity.Store[component.Position]
	Stats         *entity.Store[component.Stats]
	Factions      *entity.Store[component.FactionComp]
	AIs           *entity.Store[component.AI]
	Inventories   *entity.Store[component.Inventory]
	Anims         *entity.Store[component.AnimState]
	Appearances   *entity.Store[component.Appearance]
	Names         *entity.Store[component.Name]
	Rewards       *entity.Store[component.Reward]
	Elements      *entity.Store[component.Element]
	Races         *entity.Store[component.Race]
	StatusEffects *entity.Store[component.StatusEffects]

	Player    entity.ID
	World     *world.GameMap
	Scheduler *turn.Scheduler
	registry  *content.Registry `json:"-"`
	Resolver  *rules.Resolver   `json:"-"`
	AnimQueue *animation.Queue  `json:"-"`
	rng       rules.RNG

	PlayState  PlayState
	MessageLog []string
	Message    string
	Frame      int

	MenuOpen   bool
	MenuCursor int
	MenuItems  []MenuItem

	ConfirmQuit  bool
	ConfirmStair bool
	StairLeft    []string
	PendingDir   int

	Audio       AudioPlayer
	Debug       *debug.State
	Clock       *clock.ScaledClock
	SaveService *save.Service
}

type MenuItem struct {
	Label   string
	Enabled bool
	Kind    int
}

const (
	menuAttack    = 0
	menuCompanion = 1
	menuExamine   = 2
	menuItem      = 3
	menuWait      = 4
	menuSave      = 5
)

func RestoreGameScene(snap *save.Snapshot, reg *content.Registry, audio AudioPlayer, ss *save.Service) *GameScene {
	g := &GameScene{
		Entities:      entity.NewManager(),
		Positions:     entity.NewStore[component.Position](),
		Stats:         entity.NewStore[component.Stats](),
		Factions:      entity.NewStore[component.FactionComp](),
		AIs:           entity.NewStore[component.AI](),
		Inventories:   entity.NewStore[component.Inventory](),
		Anims:         entity.NewStore[component.AnimState](),
		Appearances:   entity.NewStore[component.Appearance](),
		Names:         entity.NewStore[component.Name](),
		Rewards:       entity.NewStore[component.Reward](),
		Elements:      entity.NewStore[component.Element](),
		Races:         entity.NewStore[component.Race](),
		StatusEffects: entity.NewStore[component.StatusEffects](),
		Scheduler:     turn.NewScheduler(),
		registry:      reg,
		AnimQueue:     animation.NewQueue(),
		rng:           rules.NewRNG(snap.MapSeed),
		Audio:         audio,
		Debug:         debug.NewState(),
		Clock:         clock.New(),
		SaveService:   ss,
	}

	g.Player = entity.ID(snap.PlayerID)
	g.Scheduler.TurnCount = snap.TurnCount

	floorDef := reg.GetFloorDef(snap.Floor)
	g.World = mapgen.Generate(snap.Floor, floorDef, reg, snap.MapSeed, g)

	for _, es := range snap.Entities {
		id := entity.ID(es.ID)
		g.Entities.CreateWithID(id)
		save.RestoreEntityComponents(es, id,
			g.Positions, g.Stats, g.Factions, g.AIs, g.Inventories,
			g.Anims, g.Appearances, g.Names, g.Rewards, g.Elements, g.Races)
	}

	save.RestoreMapExplored(g.World, snap.Explored)
	save.RestoreMapItems(g.World, snap.MapItems)

	if len(snap.MessageLog) > 0 {
		g.MessageLog = snap.MessageLog
		g.Message = snap.MessageLog[len(snap.MessageLog)-1]
	}

	g.Resolver = &rules.Resolver{Access: g, PlayerID: g.Player, RNG: g.rng}
	pPos := g.playerPos()
	if pPos != nil {
		world.UpdateVisibility(g.World, pPos.X, pPos.Y)
	}

	g.pushMessage("ロード完了！")
	return g
}

func NewGameScene(reg *content.Registry, audio AudioPlayer, ss *save.Service) *GameScene {
	g := &GameScene{
		Entities:      entity.NewManager(),
		Positions:     entity.NewStore[component.Position](),
		Stats:         entity.NewStore[component.Stats](),
		Factions:      entity.NewStore[component.FactionComp](),
		AIs:           entity.NewStore[component.AI](),
		Inventories:   entity.NewStore[component.Inventory](),
		Anims:         entity.NewStore[component.AnimState](),
		Appearances:   entity.NewStore[component.Appearance](),
		Names:         entity.NewStore[component.Name](),
		Rewards:       entity.NewStore[component.Reward](),
		Elements:      entity.NewStore[component.Element](),
		Races:         entity.NewStore[component.Race](),
		StatusEffects: entity.NewStore[component.StatusEffects](),
		Scheduler:     turn.NewScheduler(),
		registry:      reg,
		AnimQueue:     animation.NewQueue(),
		rng:           rules.NewRNG(0), // Placeholder, will be set in generateFloor
		Audio:         audio,
		Debug:         debug.NewState(),
		Clock:         clock.New(),
		SaveService:   ss,
	}
	g.Resolver = &rules.Resolver{Access: g, PlayerID: entity.ID(1), RNG: g.rng}
	g.spawnPlayer(reg)
	g.generateFloor(0)
	return g
}

func (g *GameScene) playerPos() *component.Position {
	p, _ := g.Positions.Get(g.Player)
	return p
}

func (g *GameScene) pushMessage(msg string) {
	g.Message = msg
	g.MessageLog = append(g.MessageLog, msg)
	if len(g.MessageLog) > 1000 {
		g.MessageLog = g.MessageLog[len(g.MessageLog)-1000:]
	}
}

// save.GameAccess implementation
func (g *GameScene) GetPlayerID() entity.ID     { return g.Player }
func (g *GameScene) GetFloor() int              { return g.World.Floor }
func (g *GameScene) GetTurnCount() int          { return g.Scheduler.TurnCount }
func (g *GameScene) GetMapSeed() int64          { return g.World.Seed }
func (g *GameScene) GetMessageLog() []string    { return g.MessageLog }
func (g *GameScene) GetGameMap() *world.GameMap { return g.World }
func (g *GameScene) AllEntities() []entity.ID   { return g.Entities.All() }
func (g *GameScene) GetAppearance(id entity.ID) *component.Appearance {
	a, _ := g.Appearances.Get(id)
	return a
}

// WorldAccess implementation for action system
func (g *GameScene) GetPosition(id entity.ID) *component.Position {
	p, _ := g.Positions.Get(id)
	return p
}
func (g *GameScene) GetStats(id entity.ID) *component.Stats {
	s, _ := g.Stats.Get(id)
	return s
}
func (g *GameScene) GetFaction(id entity.ID) *component.FactionComp {
	f, _ := g.Factions.Get(id)
	return f
}
func (g *GameScene) GetInventory(id entity.ID) *component.Inventory {
	inv, _ := g.Inventories.Get(id)
	return inv
}
func (g *GameScene) GetAI(id entity.ID) *component.AI {
	a, _ := g.AIs.Get(id)
	return a
}
func (g *GameScene) GetAnim(id entity.ID) *component.AnimState {
	a, _ := g.Anims.Get(id)
	return a
}
func (g *GameScene) GetName(id entity.ID) string {
	n, ok := g.Names.Get(id)
	if !ok {
		return "???"
	}
	return n.Value
}
func (g *GameScene) GetElement(id entity.ID) component.Element {
	e, ok := g.Elements.Get(id)
	if !ok {
		return component.ElementNone
	}
	return *e
}
func (g *GameScene) GetRace(id entity.ID) component.Race {
	r, ok := g.Races.Get(id)
	if !ok {
		return component.RaceHuman
	}
	return *r
}
func (g *GameScene) GetRewardXP(id entity.ID) int {
	r, ok := g.Rewards.Get(id)
	if !ok {
		return 0
	}
	return r.XP
}
func (g *GameScene) IsAlive(id entity.ID) bool {
	return g.Entities.IsAlive(id)
}
func (g *GameScene) UnitAt(x, y int) entity.ID {
	var found entity.ID
	g.Positions.Each(func(id entity.ID, pos *component.Position) {
		if pos.X == x && pos.Y == y && g.Entities.IsAlive(id) {
			found = id
		}
	})
	return found
}
func (g *GameScene) Map() *world.GameMap {
	return g.World
}
func (g *GameScene) PlayerID() entity.ID {
	return g.Player
}

// rules.EntityAccess implementation
func (g *GameScene) Destroy(id entity.ID) {
	g.Entities.Destroy(id)
}

// ai.EntityQuery implementation
func (g *GameScene) EntitiesWithFaction(f component.Faction) []entity.ID {
	var ids []entity.ID
	g.Factions.Each(func(id entity.ID, fc *component.FactionComp) {
		if fc.Faction == f && g.Entities.IsAlive(id) {
			ids = append(ids, id)
		}
	})
	return ids
}

func (g *GameScene) CheckCondition(condition, itemID string) bool {
	switch condition {
	case "unique":
		return !g.isItemPossessed(itemID)
	default:
		return true
	}
}

func (g *GameScene) RNG() rules.RNG              { return g.rng }
func (g *GameScene) Registry() *content.Registry { return g.registry }

func equippedPhyAtk(inv *component.Inventory, reg *content.Registry) int {
	if inv == nil {
		return 0
	}
	atk := 0
	for _, e := range inv.Items {
		if e.Equipped {
			def, ok := reg.GetItemDef(e.DefID)
			if ok {
				atk += def.PhyAtk
			}
		}
	}
	return atk
}

func equippedPhyDef(inv *component.Inventory, reg *content.Registry) int {
	if inv == nil {
		return 0
	}
	d := 0
	for _, e := range inv.Items {
		if e.Equipped {
			def, ok := reg.GetItemDef(e.DefID)
			if ok {
				d += def.PhyDef
			}
		}
	}
	return d
}

func (g *GameScene) hasTreasure() bool {
	inv := g.GetInventory(g.Player)
	if inv == nil {
		return false
	}
	for _, entry := range inv.Items {
		def, ok := g.registry.GetItemDef(entry.DefID)
		if ok && def.FloorBound {
			return true
		}
	}
	return false
}

func (g *GameScene) calcScore() int {
	stats := g.GetStats(g.Player)
	if stats == nil {
		return 0
	}
	base := stats.HP*100 - g.Scheduler.TurnCount
	if base < 0 {
		base = 0
	}
	return base * g.registry.FloorCount()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
