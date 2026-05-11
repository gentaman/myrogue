package app

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gentaman/myrogue/internal/animation"
	"github.com/gentaman/myrogue/internal/core/action"
	"github.com/gentaman/myrogue/internal/core/ai"
	"github.com/gentaman/myrogue/internal/core/combat"
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
	"github.com/gentaman/myrogue/internal/core/entity"
	"github.com/gentaman/myrogue/internal/core/rules"
	"github.com/gentaman/myrogue/internal/core/turn"
	"github.com/gentaman/myrogue/internal/core/world"
	"github.com/gentaman/myrogue/internal/mapgen"
)

type PlayState int

const (
	StatePlaying PlayState = iota
	StateWin
	StateDead
)

type GameScene struct {
	Entities    *entity.Manager
	Positions   *entity.Store[component.Position]
	Stats       *entity.Store[component.Stats]
	Factions    *entity.Store[component.FactionComp]
	AIs         *entity.Store[component.AI]
	Inventories *entity.Store[component.Inventory]
	Anims       *entity.Store[component.AnimState]
	Appearances *entity.Store[component.Appearance]
	Names       *entity.Store[component.Name]
	Rewards     *entity.Store[component.Reward]
	Elements    *entity.Store[component.Element]
	Races       *entity.Store[component.Race]

	Player    entity.ID
	World     *world.GameMap
	Scheduler *turn.Scheduler
	Registry  *content.Registry
	Resolver  *rules.Resolver
	AnimQueue *animation.Queue

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

	Audio AudioPlayer
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
)

func NewGameScene(reg *content.Registry, audio AudioPlayer) *GameScene {
	g := &GameScene{
		Entities:    entity.NewManager(),
		Positions:   entity.NewStore[component.Position](),
		Stats:       entity.NewStore[component.Stats](),
		Factions:    entity.NewStore[component.FactionComp](),
		AIs:         entity.NewStore[component.AI](),
		Inventories: entity.NewStore[component.Inventory](),
		Anims:       entity.NewStore[component.AnimState](),
		Appearances: entity.NewStore[component.Appearance](),
		Names:       entity.NewStore[component.Name](),
		Rewards:     entity.NewStore[component.Reward](),
		Elements:    entity.NewStore[component.Element](),
		Races:       entity.NewStore[component.Race](),
		Scheduler:   turn.NewScheduler(),
		Registry:    reg,
		AnimQueue:   animation.NewQueue(),
		Audio:       audio,
	}
	g.Resolver = &rules.Resolver{Access: g, PlayerID: entity.ID(1)}
	g.spawnPlayer(reg)
	g.generateFloor(0)
	return g
}

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
}

func (g *GameScene) generateFloor(floor int) {
	floorDef := g.Registry.GetFloorDef(floor)
	seed := time.Now().UnixNano()
	g.World = mapgen.Generate(floor, floorDef, g.Registry, seed)

	if len(g.World.Rooms) > 0 {
		r := g.World.Rooms[0]
		pos := g.playerPos()
		pos.X = r.CenterX()
		pos.Y = r.CenterY()
	}

	g.spawnInitialEnemies()
	g.spawnCompanion("dog")
	world.UpdateVisibility(g.World, g.playerPos().X, g.playerPos().Y)
}

func (g *GameScene) spawnInitialEnemies() {
	floorDef := g.Registry.GetFloorDef(g.World.Floor)
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
		def, ok := g.Registry.GetEnemyDef(defID)
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
	return id
}

func (g *GameScene) spawnCompanion(defID string) {
	def, ok := g.Registry.GetCompanionDef(defID)
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
			return
		}
	}
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

func (g *GameScene) isAnyAnimating() bool {
	anyAnim := false
	g.Anims.Each(func(id entity.ID, a *component.AnimState) {
		if a.AttackAnim > 0 || a.DamageAnim > 0 {
			anyAnim = true
		}
	})
	return anyAnim || g.AnimQueue.IsPlaying()
}

func (g *GameScene) tickAnimations() {
	g.Anims.Each(func(id entity.ID, a *component.AnimState) {
		if a.AttackAnim > 0 {
			a.AttackAnim--
		}
		if a.DamageAnim > 0 {
			a.DamageAnim--
		}
	})
	g.AnimQueue.Tick()
}

func (g *GameScene) processEvents(events []action.Event) {
	for _, ev := range events {
		switch e := ev.(type) {
		case action.EventLog:
			g.pushMessage(e.Text)
		case action.EventSFX:
			if g.Audio != nil {
				g.Audio.PlaySFX(e.ID)
			}
		case action.EventDeath:
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
		case action.EventXP:
			cascaded := g.Resolver.ProcessXP(e)
			if e.Entity == g.Player {
				g.pushMessage(fmt.Sprintf("%d の経験値を獲得！", e.Amount))
			}
			g.processEvents(cascaded)
		case action.EventLevelUp:
			if e.Entity == g.Player {
				g.pushMessage(fmt.Sprintf("レベル %d に上がった！", e.NewLevel))
			}
		case action.EventAttack:
			anim, _ := g.Anims.Get(e.Defender)
			if anim != nil {
				anim.DamageAnim = component.DamageAnimFrames
			}
		case action.EventDrop:
			g.World.Items = append(g.World.Items, world.MapItem{X: e.X, Y: e.Y, Inventory: e.Items})
		case action.EventProjectile:
			g.AnimQueue.Add(animation.Projectile{
				StartX: e.StartX, StartY: e.StartY,
				EndX: e.EndX, EndY: e.EndY,
				TotalFrames: e.TotalFrames,
				ColorHex:    e.ColorHex,
				IsFlash:     e.IsFlash,
			})
		}
	}
}

func (g *GameScene) resolveCombat(attackerID, defenderID entity.ID) {
	atkStats := g.GetStats(attackerID)
	defStats := g.GetStats(defenderID)
	if atkStats == nil || defStats == nil {
		return
	}

	atkInv := g.GetInventory(attackerID)
	defInv := g.GetInventory(defenderID)

	atk := &combat.Combatant{
		ID:      attackerID,
		Name:    g.GetName(attackerID),
		Stats:   atkStats,
		Element: g.GetElement(attackerID),
		Race:    g.GetRace(attackerID),
		PhyAtk:  equippedPhyAtk(atkInv, g.Registry),
		PhyDef:  equippedPhyDef(atkInv, g.Registry),
	}
	def2 := &combat.Combatant{
		ID:      defenderID,
		Name:    g.GetName(defenderID),
		Stats:   defStats,
		Element: g.GetElement(defenderID),
		Race:    g.GetRace(defenderID),
		PhyAtk:  equippedPhyAtk(defInv, g.Registry),
		PhyDef:  equippedPhyDef(defInv, g.Registry),
	}

	events := combat.ResolveCombat(atk, def2, combat.CombatTypePhysical, 0, atk.Element, g.Player)
	g.processEvents(events)

	// XP on kill
	if defStats.HP <= 0 && attackerID != defenderID {
		xp := g.GetRewardXP(defenderID)
		if xp > 0 {
			g.processEvents([]action.Event{action.EventXP{Entity: attackerID, Amount: xp}})
		}
	}
}

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

func (g *GameScene) Update(input InputState) Scene {
	g.Frame++

	if g.isAnyAnimating() {
		g.tickAnimations()
		return nil
	}

	if g.PlayState == StateWin || g.PlayState == StateDead {
		if input.Restart {
			return NewGameScene(g.Registry, g.Audio)
		}
		if input.Cancel {
			return NewTitleSceneWithDeps(g.Registry, g.Audio)
		}
		return nil
	}

	if g.ConfirmQuit {
		if input.Cancel || input.No {
			g.ConfirmQuit = false
		} else if input.Confirm || input.Yes {
			return NewTitleSceneWithDeps(g.Registry, g.Audio)
		}
		return nil
	}

	if g.ConfirmStair {
		if input.Cancel || input.No {
			g.ConfirmStair = false
		} else if input.Confirm || input.Yes {
			g.ConfirmStair = false
			// TODO: floor change
		}
		return nil
	}

	switch g.Scheduler.Phase {
	case turn.PhasePlayerInput:
		return g.updatePlayerTurn(input)
	case turn.PhaseCompanionAct:
		g.updateCompanionTurns()
	case turn.PhaseEnemyAct:
		g.updateEnemyTurns()
	}
	return nil
}

func (g *GameScene) updatePlayerTurn(input InputState) Scene {
	if input.Cancel {
		if g.MenuOpen {
			g.MenuOpen = false
			return nil
		}
		g.ConfirmQuit = true
		return nil
	}

	if input.Menu {
		g.MenuOpen = !g.MenuOpen
		if g.MenuOpen {
			g.MenuCursor = 0
			g.buildMenu()
		}
		return nil
	}

	if g.MenuOpen {
		if input.Up {
			g.MenuCursor--
			if g.MenuCursor < 0 {
				g.MenuCursor = len(g.MenuItems) - 1
			}
		}
		if input.Down {
			g.MenuCursor++
			if g.MenuCursor >= len(g.MenuItems) {
				g.MenuCursor = 0
			}
		}
		if input.Confirm {
			g.execMenuItem()
		}
		return nil
	}

	if input.DirPressed {
		g.handleMovement(input.Dir)
	}
	return nil
}

func (g *GameScene) handleMovement(dir int) {
	var dx, dy int
	switch dir {
	case 0:
		dy = -1
	case 1:
		dy = 1
	case 2:
		dx = -1
	case 3:
		dx = 1
	}
	pPos := g.playerPos()
	pPos.Dir = component.DirFromDelta(dx, dy)
	nx, ny := pPos.X+dx, pPos.Y+dy
	if !g.World.InBounds(nx, ny) || !g.World.IsPassable(nx, ny) {
		return
	}

	g.Scheduler.IncrementTurn()
	occupant := g.UnitAt(nx, ny)
	if occupant != entity.InvalidID {
		faction, _ := g.Factions.Get(occupant)
		if faction != nil && faction.Faction == component.FactionAlly {
			// swap with companion
			oPos := g.GetPosition(occupant)
			oPos.X, oPos.Y = pPos.X, pPos.Y
			pPos.X, pPos.Y = nx, ny
		} else {
			// attack
			anim, _ := g.Anims.Get(g.Player)
			if anim != nil {
				anim.AttackAnim = component.AttackAnimFrames
			}
			g.resolveCombat(g.Player, occupant)
			if g.Audio != nil {
				g.Audio.PlaySFX("hit")
			}
		}
	} else {
		pPos.X, pPos.Y = nx, ny
	}
	world.UpdateVisibility(g.World, pPos.X, pPos.Y)
	g.Scheduler.StartCompanionPhase()
}

func (g *GameScene) updateCompanionTurns() {
	allies := g.EntitiesWithFaction(component.FactionAlly)
	if g.Scheduler.ActiveIdx >= len(allies) {
		g.Scheduler.StartEnemyPhase()
		return
	}
	id := allies[g.Scheduler.ActiveIdx]
	g.runAI(id, &ai.AllyBrain{})
	g.Scheduler.AdvanceIdx()
}

func (g *GameScene) updateEnemyTurns() {
	enemies := g.EntitiesWithFaction(component.FactionEnemy)
	if g.Scheduler.ActiveIdx >= len(enemies) {
		g.Scheduler.StartPlayerPhase()
		// MP regen
		stats := g.GetStats(g.Player)
		if stats != nil && stats.MP < stats.MaxMP {
			stats.MP++
		}
		return
	}
	id := enemies[g.Scheduler.ActiveIdx]
	g.runAI(id, &ai.HostileBrain{})
	g.Scheduler.AdvanceIdx()
}

func (g *GameScene) runAI(id entity.ID, brain ai.Brain) {
	pos := g.GetPosition(id)
	aiComp := g.GetAI(id)
	if pos == nil {
		return
	}
	if aiComp == nil {
		aiComp = &component.AI{}
	}

	ctx := &ai.DecisionContext{
		Self:    id,
		SelfPos: pos,
		SelfAI:  aiComp,
		World:   g.World,
		Query:   g,
	}
	act := brain.Decide(ctx)
	if act == nil {
		return
	}
	switch a := act.(type) {
	case *action.MoveAction:
		events := a.Execute(id, g)
		g.processEvents(events)
	case *action.AttackAction:
		pos.Dir = component.DirFromDelta(a.TargetX-pos.X, a.TargetY-pos.Y)
		anim, _ := g.Anims.Get(id)
		if anim != nil {
			anim.AttackAnim = component.AttackAnimFrames
		}
		target := g.UnitAt(a.TargetX, a.TargetY)
		if target != entity.InvalidID {
			g.resolveCombat(id, target)
		}
	case *action.WaitAction:
		// do nothing
	}
}

func (g *GameScene) buildMenu() {
	g.MenuItems = nil
	pPos := g.playerPos()
	dx, dy := pPos.Dir.Delta()
	fx, fy := pPos.X+dx, pPos.Y+dy

	hasEnemy := false
	occ := g.UnitAt(fx, fy)
	if occ != entity.InvalidID {
		f, _ := g.Factions.Get(occ)
		if f != nil && f.Faction == component.FactionEnemy {
			hasEnemy = true
		}
	}
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "正面を攻撃", Enabled: hasEnemy, Kind: menuAttack})

	hasCompanion := false
	if occ != entity.InvalidID {
		f, _ := g.Factions.Get(occ)
		if f != nil && f.Faction == component.FactionAlly {
			hasCompanion = true
		}
	}
	compLabel := "仲間に指示を出す"
	if hasCompanion {
		compLabel = g.GetName(occ) + "に指示を出す"
	}
	g.MenuItems = append(g.MenuItems, MenuItem{Label: compLabel, Enabled: hasCompanion, Kind: menuCompanion})

	examineLabel := "足元を調べる"
	g.MenuItems = append(g.MenuItems, MenuItem{Label: examineLabel, Enabled: true, Kind: menuExamine})

	inv := g.GetInventory(g.Player)
	hasItems := inv != nil && len(inv.Items) > 0
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "道具を使う", Enabled: hasItems, Kind: menuItem})
	g.MenuItems = append(g.MenuItems, MenuItem{Label: "その場で待機", Enabled: true, Kind: menuWait})
}

func (g *GameScene) execMenuItem() {
	if g.MenuCursor >= len(g.MenuItems) {
		return
	}
	item := g.MenuItems[g.MenuCursor]
	if !item.Enabled {
		return
	}
	g.MenuOpen = false
	switch item.Kind {
	case menuAttack:
		pPos := g.playerPos()
		dx, dy := pPos.Dir.Delta()
		fx, fy := pPos.X+dx, pPos.Y+dy
		g.Scheduler.IncrementTurn()
		anim, _ := g.Anims.Get(g.Player)
		if anim != nil {
			anim.AttackAnim = component.AttackAnimFrames
		}
		target := g.UnitAt(fx, fy)
		if target != entity.InvalidID {
			g.resolveCombat(g.Player, target)
		}
		g.Scheduler.StartCompanionPhase()
	case menuWait:
		g.Scheduler.IncrementTurn()
		g.pushMessage("待機した。")
		g.Scheduler.StartCompanionPhase()
	case menuExamine:
		g.Scheduler.IncrementTurn()
		g.pushMessage("特に何もない。")
		g.Scheduler.StartCompanionPhase()
	}
}

func (g *GameScene) cameraOffset() (float64, float64) {
	pPos := g.playerPos()
	cx := float64(pPos.X*TileSize+TileSize/2) - float64(ScreenWidth)/2
	cy := float64(pPos.Y*TileSize+TileSize/2) - float64(GameplayAreaHeight)/2
	maxX := float64(world.MapWidth*TileSize - ScreenWidth)
	maxY := float64(world.MapHeight*TileSize - GameplayAreaHeight)
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	if cx > maxX {
		cx = maxX
	}
	if cy > maxY {
		cy = maxY
	}
	return cx, cy
}

func (g *GameScene) charAnimOffset(attackAnim int, dir component.Dir) (int, int) {
	if attackAnim > 0 {
		shift := 1
		if attackAnim > component.AttackAnimFrames/2 {
			shift = 3
		}
		dx, dy := dir.Delta()
		return dx * shift, dy * shift
	}
	return 0, 0
}

func (g *GameScene) Draw(r Renderer) {
	r.Clear(20, 20, 20, 255)
	g.drawWorld(r)
	g.drawUI(r)
	g.drawOverlays(r)
}

func (g *GameScene) drawWorld(r Renderer) {
	camX, camY := g.cameraOffset()

	for x := 0; x < world.MapWidth; x++ {
		for y := 0; y < world.MapHeight; y++ {
			if !g.World.Explored[x][y] {
				continue
			}
			vis := g.World.Visible[x][y]
			var clr Color
			if vis {
				switch g.World.Tiles[x][y] {
				case world.Wall:
					clr = Color{100, 100, 110, 255}
				case world.Floor:
					clr = Color{60, 60, 65, 255}
				case world.Stairs:
					clr = Color{0, 255, 255, 255}
				case world.StairsDown:
					clr = Color{255, 140, 0, 255}
				case world.StairsUp:
					clr = Color{0, 220, 180, 255}
				}
			} else {
				switch g.World.Tiles[x][y] {
				case world.Wall:
					clr = Color{50, 50, 55, 255}
				case world.Floor:
					clr = Color{28, 28, 30, 255}
				case world.Stairs:
					clr = Color{0, 120, 120, 255}
				case world.StairsDown:
					clr = Color{140, 70, 0, 255}
				case world.StairsUp:
					clr = Color{0, 110, 90, 255}
				}
			}
			sx := float64(x*TileSize+1) - camX
			sy := float64(y*TileSize+1) - camY
			if sx > float64(ScreenWidth) || sy > float64(GameplayAreaHeight) || sx < -float64(TileSize) || sy < -float64(TileSize) {
				continue
			}
			r.DrawRect(int(sx), int(sy), TileSize-2, TileSize-2, clr)
		}
	}

	for _, it := range g.World.Items {
		if !g.World.Explored[it.X][it.Y] {
			continue
		}
		var clr Color
		if len(it.Inventory) > 1 {
			clr = Color{180, 130, 40, 255}
		} else if len(it.Inventory) == 1 {
			clr = Color{200, 200, 100, 255}
		} else {
			continue
		}
		if !g.World.Visible[it.X][it.Y] {
			clr = Color{clr[0] / 2, clr[1] / 2, clr[2] / 2, 255}
		}
		ix := float64(it.X*TileSize+8) - camX
		iy := float64(it.Y*TileSize+8) - camY
		r.DrawRect(int(ix), int(iy), TileSize-16, TileSize-16, clr)
	}

	// Draw entities
	g.Factions.Each(func(id entity.ID, fc *component.FactionComp) {
		if !g.Entities.IsAlive(id) {
			return
		}
		pos := g.GetPosition(id)
		if pos == nil {
			return
		}
		if !g.World.Visible[pos.X][pos.Y] && id != g.Player {
			return
		}

		anim := g.GetAnim(id)
		if anim != nil && anim.DamageAnim > 0 && (anim.DamageAnim/4)%2 == 0 {
			return
		}

		attackAnim := 0
		if anim != nil {
			attackAnim = anim.AttackAnim
		}
		ox, oy := g.charAnimOffset(attackAnim, pos.Dir)

		var clr Color
		switch fc.Faction {
		case component.FactionPlayer:
			clr = Color{255, 255, 255, 255}
		case component.FactionAlly:
			clr = Color{100, 200, 100, 255}
		case component.FactionEnemy:
			clr = Color{200, 80, 80, 255}
		}

		app := g.Appearances
		if app != nil {
			a, ok := app.Get(id)
			if ok && a.ColorHex != "" {
				clr = hexToColor(a.ColorHex)
			}
		}

		sx := float64(pos.X*TileSize+ox*2) - camX
		sy := float64(pos.Y*TileSize+oy*2) - camY
		r.DrawRect(int(sx), int(sy), TileSize-1, TileSize-1, clr)

		// Direction dot
		var dotClr Color
		switch fc.Faction {
		case component.FactionPlayer:
			dotClr = Color{255, 255, 255, 255}
		case component.FactionAlly:
			alpha := uint8(180 + 40*((g.Frame/30)%2))
			dotClr = Color{100, 255, 100, alpha}
		case component.FactionEnemy:
			aiComp := g.GetAI(id)
			if aiComp != nil && aiComp.State > 0 {
				if (g.Frame/5)%2 == 0 {
					dotClr = Color{255, 50, 50, 255}
				} else {
					dotClr = Color{255, 200, 200, 255}
				}
			} else {
				alpha := uint8(180 + 40*((g.Frame/30)%2))
				dotClr = Color{255, 255, 255, alpha}
			}
		}
		g.drawDirDot(r, pos.X, pos.Y, pos.Dir, dotClr, camX, camY)
	})

	// Projectiles
	for _, p := range g.AnimQueue.Projectiles {
		if p.IsFlash {
			alpha := uint8(200)
			if (p.Frame/4)%2 == 0 {
				alpha = 100
			}
			clr := hexToColor(p.ColorHex)
			clr[3] = alpha
			px := p.EndX - camX
			py := p.EndY - camY
			r.DrawRect(int(px), int(py), TileSize, TileSize, clr)
			continue
		}
		t := float64(p.Frame) / float64(p.TotalFrames)
		curX := p.StartX + (p.EndX-p.StartX)*t
		curY := p.StartY + (p.EndY-p.StartY)*t
		clr := hexToColor(p.ColorHex)
		r.DrawRect(int(curX-camX-4), int(curY-camY-4), 8, 8, clr)
	}
}

func (g *GameScene) drawDirDot(r Renderer, tileX, tileY int, d component.Dir, clr Color, camX, camY float64) {
	const dotSize = 6
	const half = (TileSize - 1) / 2
	ox, oy := 0, 0
	switch d {
	case component.DirUp:
		ox, oy = half-dotSize/2, 4
	case component.DirDown:
		ox, oy = half-dotSize/2, TileSize-1-dotSize-4
	case component.DirLeft:
		ox, oy = 4, half-dotSize/2
	case component.DirRight:
		ox, oy = TileSize-1-dotSize-4, half-dotSize/2
	}
	sx := float64(tileX*TileSize+ox) - camX
	sy := float64(tileY*TileSize+oy) - camY
	r.DrawRect(int(sx), int(sy), dotSize, dotSize, clr)
}

func (g *GameScene) drawUI(r Renderer) {
	r.DrawRect(0, GameplayAreaHeight, ScreenWidth, ScreenHeight-GameplayAreaHeight, Color{0, 0, 0, 255})
	statusY := GameplayAreaHeight + 8

	hpClr := Color{200, 200, 200, 255}
	stats := g.GetStats(g.Player)
	if stats != nil && stats.HP <= stats.MaxHP/4 {
		hpClr = Color{255, 80, 80, 255}
	}

	r.DrawText(fmt.Sprintf("フロア: %d  ターン: %d", g.World.Floor+1, g.Scheduler.TurnCount), 12, 8, statusY, Color{200, 200, 200, 255}, false)
	if stats != nil {
		r.DrawText(fmt.Sprintf("Lv: %d  XP: %d / %d", stats.Level, stats.XP, stats.XPToNext), 12, 160, statusY, Color{200, 200, 200, 255}, false)
		r.DrawText(fmt.Sprintf("HP: %d / %d  MP: %d / %d", stats.HP, stats.MaxHP, stats.MP, stats.MaxMP), 12, 8, statusY+16, hpClr, false)
		atk := 1 + stats.Str + equippedPhyAtk(g.GetInventory(g.Player), g.Registry)
		def := equippedPhyDef(g.GetInventory(g.Player), g.Registry) + stats.Vit
		r.DrawText(fmt.Sprintf("ATK: %d  DEF: %d", atk, def), 12, 160, statusY+16, Color{200, 200, 200, 255}, false)
		r.DrawText(fmt.Sprintf("Str:%d Wis:%d Fai:%d Vit:%d Agi:%d Luk:%d", stats.Str, stats.Wis, stats.Fai, stats.Vit, stats.Agi, stats.Luk), 12, 8, statusY+32, Color{150, 150, 150, 255}, false)
	}
	r.DrawText(g.Message, 14, 8, statusY+48, Color{255, 255, 255, 255}, false)
	r.DrawText("H: ヘルプ", 12, ScreenWidth-80, statusY, Color{100, 100, 100, 255}, false)
}

func (g *GameScene) drawOverlays(r Renderer) {
	if g.MenuOpen {
		const (
			rowH      = 32
			padX      = 48
			padRight  = 16
			headerH   = 44
			footerH   = 28
			minPanelW = 240
		)
		panelW := minPanelW
		for _, item := range g.MenuItems {
			w := r.MeasureText(item.Label, 14) + padX + padRight
			if w > panelW {
				panelW = w
			}
		}
		panelH := headerH + len(g.MenuItems)*rowH + footerH
		if panelH > ScreenHeight-40 {
			panelH = ScreenHeight - 40
		}
		visibleRows := (panelH - headerH - footerH) / rowH
		if visibleRows < 1 {
			visibleRows = 1
		}
		scrollOffset := g.MenuCursor - visibleRows + 1
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		if g.MenuCursor < scrollOffset {
			scrollOffset = g.MenuCursor
		}
		panelX, panelY := (ScreenWidth-panelW)/2, (ScreenHeight-panelH)/2
		r.DrawPanel(panelX, panelY, panelW, panelH, Color{20, 20, 50, 255}, Color{100, 150, 255, 255})
		r.DrawText("アクション", 14, ScreenWidth/2, panelY+12, Color{255, 220, 100, 255}, true)
		for i, item := range g.MenuItems {
			row := i - scrollOffset
			if row < 0 || row >= visibleRows {
				continue
			}
			y := panelY + headerH + row*rowH
			clr := Color{180, 180, 180, 255}
			if !item.Enabled {
				clr = Color{80, 80, 80, 255}
			}
			if i == g.MenuCursor {
				r.DrawRect(panelX+4, y-4, panelW-8, 26, Color{60, 60, 120, 255})
				clr = Color{255, 255, 255, 255}
			}
			r.DrawText(item.Label, 14, panelX+30, y, clr, false)
		}
		r.DrawText("W/S: 選択  Enter: 決定  X/Esc: 閉じる", 12, ScreenWidth/2, panelY+panelH-20, Color{120, 120, 120, 255}, true)
	}

	if g.ConfirmStair {
		const (
			dW = 340
			dH = 100
			dX = (ScreenWidth - dW) / 2
			dY = (ScreenHeight - dH) / 2
		)
		r.DrawPanel(dX, dY, dW, dH, Color{30, 30, 20, 255}, Color{200, 180, 100, 255})
		r.DrawText("はぐれる仲間がいます！", 14, ScreenWidth/2, dY+14, Color{255, 220, 100, 255}, true)
		names := ""
		for i, n := range g.StairLeft {
			if i > 0 {
				names += "、"
			}
			names += n
		}
		r.DrawText(names+" は離れすぎている", 12, ScreenWidth/2, dY+40, Color{220, 220, 220, 255}, true)
		r.DrawText("それでも移動しますか？", 12, ScreenWidth/2, dY+58, Color{200, 200, 200, 255}, true)
		r.DrawText("Enter/Y: はい    Esc/N: いいえ", 12, ScreenWidth/2, dY+80, Color{180, 180, 180, 255}, true)
	}

	if g.ConfirmQuit {
		const (
			dW = 300
			dH = 80
			dX = (ScreenWidth - dW) / 2
			dY = (ScreenHeight - dH) / 2
		)
		r.DrawPanel(dX, dY, dW, dH, Color{30, 20, 20, 255}, Color{200, 100, 100, 255})
		r.DrawText("タイトルへ戻りますか？", 14, ScreenWidth/2, dY+14, Color{255, 220, 220, 255}, true)
		r.DrawText("Enter/Y: はい    Esc/N: いいえ", 12, ScreenWidth/2, dY+46, Color{180, 180, 180, 255}, true)
	}

	if g.PlayState == StateWin || g.PlayState == StateDead {
		const (
			dW = 360
			dH = 120
			dX = (ScreenWidth - dW) / 2
			dY = (ScreenHeight - dH) / 2
		)
		var pClr, bClr Color
		var title, sub string
		if g.PlayState == StateWin {
			pClr = Color{20, 40, 20, 255}
			bClr = Color{100, 255, 100, 255}
			title = "--- BEAT THE GAME ---"
			sub = "おめでとうございます！"
		} else {
			pClr = Color{40, 20, 20, 255}
			bClr = Color{255, 100, 100, 255}
			title = "--- GAME OVER ---"
			sub = "あなたは力尽きた..."
		}
		r.DrawPanel(dX, dY, dW, dH, pClr, bClr)
		r.DrawText(title, 14, ScreenWidth/2, dY+20, bClr, true)
		r.DrawText(sub, 12, ScreenWidth/2, dY+55, Color{220, 220, 220, 255}, true)
		r.DrawText("R: 再挑戦    Esc: タイトルへ", 12, ScreenWidth/2, dY+dH-25, Color{180, 180, 180, 255}, true)
	}
}

func hexToColor(hex string) Color {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return Color{255, 255, 255, 255}
	}
	parseHex := func(s string) uint8 {
		v := 0
		for _, c := range s {
			v *= 16
			switch {
			case c >= '0' && c <= '9':
				v += int(c - '0')
			case c >= 'a' && c <= 'f':
				v += int(c-'a') + 10
			case c >= 'A' && c <= 'F':
				v += int(c-'A') + 10
			}
		}
		return uint8(v)
	}
	return Color{parseHex(hex[0:2]), parseHex(hex[2:4]), parseHex(hex[4:6]), 255}
}
