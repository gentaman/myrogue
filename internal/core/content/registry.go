package content

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gentaman/myrogue/internal/core/component"
)

type Registry struct {
	Players           []ActorDef
	PlayerTemplate    ActorDef // プレイヤーの初期状態テンプレート
	SelectedCompanion string   // 選択された従者のID
	Enemies           []ActorDef
	RaceBonuses       map[component.Race][6]int
	Companions        []ActorDef
	Items             []ItemDef
	Skills            []SkillDef
	Floors            []FloorDef

	PlayerIDMap    map[string]int
	EnemyIDMap     map[string]int
	CompanionIDMap map[string]int
	ItemIDMap      map[string]int
	SkillIDMap     map[string]int

	MaxCarryWeight int
}

func NewRegistry() *Registry {
	return &Registry{
		PlayerIDMap:    make(map[string]int),
		EnemyIDMap:     make(map[string]int),
		CompanionIDMap: make(map[string]int),
		ItemIDMap:      make(map[string]int),
		SkillIDMap:     make(map[string]int),
	}
}

func (r *Registry) LoadAll(players, enemies, companions, items, floors, raceBonuses []byte, optSkills ...[]byte) error {
	var err error
	r.Players, r.PlayerIDMap, err = loadActors(players)
	if err != nil {
		return fmt.Errorf("players: %w", err)
	}
	r.Enemies, r.EnemyIDMap, err = loadActors(enemies)
	if err != nil {
		return fmt.Errorf("enemies: %w", err)
	}
	r.Companions, r.CompanionIDMap, err = loadActors(companions)
	if err != nil {
		return fmt.Errorf("companions: %w", err)
	}
	r.RaceBonuses, err = loadRaceBonuses(raceBonuses)
	if err != nil {
		return fmt.Errorf("race bonuses: %w", err)
	}
	r.Items, r.ItemIDMap, err = loadItems(items)
	if err != nil {
		return fmt.Errorf("items: %w", err)
	}
	if err := json.Unmarshal(floors, &r.Floors); err != nil {
		return fmt.Errorf("floors: %w", err)
	}
	if len(optSkills) > 0 && optSkills[0] != nil {
		r.Skills, r.SkillIDMap, err = loadSkills(optSkills[0])
		if err != nil {
			return fmt.Errorf("skills: %w", err)
		}
	}
	if len(r.Players) > 0 {
		r.MaxCarryWeight = r.Players[0].MaxCarryWeight
		r.PlayerTemplate = r.Players[0] // テンプレートとして保存
	}

	if err := r.Validate(); err != nil {
		return fmt.Errorf("registry validation failed: %w", err)
	}

	return nil
}

func (r *Registry) Validate() error {
	var errs []string

	// Actors
	validateActor := func(category string, actors []ActorDef) {
		for _, a := range actors {
			if a.ID == "" {
				errs = append(errs, fmt.Sprintf("%s: empty ID", category))
			}
			if a.HP <= 0 {
				errs = append(errs, fmt.Sprintf("%s %s: HP must be positive", category, a.ID))
			}
			if a.Rarity < 0 {
				errs = append(errs, fmt.Sprintf("%s %s: Rarity must be non-negative", category, a.ID))
			}
		}
	}
	validateActor("player", r.Players)
	validateActor("enemy", r.Enemies)
	validateActor("companion", r.Companions)

	// Items
	for _, it := range r.Items {
		if it.ID == "" {
			errs = append(errs, "item: empty ID")
		}
		if it.Weight < 0 {
			errs = append(errs, fmt.Sprintf("item %s: Weight must be non-negative", it.ID))
		}
		if it.Rarity < 0 {
			errs = append(errs, fmt.Sprintf("item %s: Rarity must be non-negative", it.ID))
		}
	}

	// Skills
	for _, sk := range r.Skills {
		if sk.ID == "" {
			errs = append(errs, "skill: empty ID")
		}
		if sk.MPCost < 0 {
			errs = append(errs, fmt.Sprintf("skill %s: MPCost must be non-negative", sk.ID))
		}
	}

	// Floors
	for _, f := range r.Floors {
		if f.MinRooms <= 0 || f.MaxRooms < f.MinRooms {
			errs = append(errs, fmt.Sprintf("floor %d: invalid room count range (%d-%d)", f.Floor, f.MinRooms, f.MaxRooms))
		}
		for _, e := range f.EnemyPool {
			if _, ok := r.EnemyIDMap[e.ID]; !ok {
				errs = append(errs, fmt.Sprintf("floor %d: enemy pool contains unknown ID %s", f.Floor, e.ID))
			}
			if e.Weight <= 0 {
				errs = append(errs, fmt.Sprintf("floor %d: enemy %s weight must be positive", f.Floor, e.ID))
			}
		}
		for _, i := range f.ItemPool {
			if _, ok := r.ItemIDMap[i.ID]; !ok {
				errs = append(errs, fmt.Sprintf("floor %d: item pool contains unknown ID %s", f.Floor, i.ID))
			}
			if i.Weight <= 0 {
				errs = append(errs, fmt.Sprintf("floor %d: item %s weight must be positive", f.Floor, i.ID))
			}
		}
		for _, ci := range f.ConditionalItems {
			if _, ok := r.ItemIDMap[ci.ID]; !ok {
				errs = append(errs, fmt.Sprintf("floor %d: conditional items contains unknown ID %s", f.Floor, ci.ID))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func loadActors(data []byte) ([]ActorDef, map[string]int, error) {
	var raws []rawActor
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, nil, err
	}
	defs := make([]ActorDef, len(raws))
	idMap := make(map[string]int)
	for i, raw := range raws {
		if _, exists := idMap[raw.ID]; exists {
			return nil, nil, fmt.Errorf("duplicate ID: %s", raw.ID)
		}
		idMap[raw.ID] = i
		defs[i] = ActorDef{
			ID:                raw.ID,
			Name:              raw.Name,
			HP:                raw.HP,
			MP:                raw.MP,
			MaxCarryWeight:    raw.MaxCarryWeight,
			Element:           component.ElementFromString(raw.Element),
			Race:              component.RaceFromString(raw.Race),
			Personality:       component.PersonalityFromString(raw.Personality),
			NeutralThreshold:  raw.NeutralThreshold,
			FriendlyThreshold: raw.FriendlyThreshold,
			FriendlyFire:      raw.FriendlyFire,
			XP:                raw.XP,
			Rarity:            raw.Rarity,
			ColorHex:          raw.Color,
			Str:               raw.Str,
			Wis:               raw.Wis,
			Fai:               raw.Fai,
			Vit:               raw.Vit,
			Agi:               raw.Agi,
			Luk:               raw.Luk,
		}
	}
	return defs, idMap, nil
}

func loadItems(data []byte) ([]ItemDef, map[string]int, error) {
	var raws []rawItem
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, nil, err
	}
	defs := make([]ItemDef, len(raws))
	idMap := make(map[string]int)
	for i, raw := range raws {
		if _, exists := idMap[raw.ID]; exists {
			return nil, nil, fmt.Errorf("duplicate ID: %s", raw.ID)
		}
		idMap[raw.ID] = i
		var slot component.EquipSlot
		var atk, def int
		if raw.Effect != nil && raw.Effect.Type == "equip" {
			switch raw.Effect.EquipType {
			case "weapon":
				slot = component.SlotWeapon
			case "shield":
				slot = component.SlotShield
			case "armor":
				slot = component.SlotArmor
			}
			atk = raw.Effect.Atk
			def = raw.Effect.Def
		}
		defs[i] = ItemDef{
			ID:         raw.ID,
			Name:       raw.Name,
			Desc:       raw.Desc,
			EffectDesc: raw.EffectDesc,
			Weight:     raw.Weight,
			Rarity:     raw.Rarity,
			ColorHex:   raw.Color,
			Durability: raw.Durability,
			FloorBound: raw.FloorBound,
			EquipSlot:  slot,
			PhyAtk:     atk,
			PhyDef:     def,
			Effect:     raw.Effect,
			Condition:  raw.CanUse,
		}
	}
	return defs, idMap, nil
}

func (r *Registry) FloorCount() int {
	return len(r.Floors)
}

func (r *Registry) GetFloorDef(floor int) *FloorDef {
	if floor < 0 || floor >= len(r.Floors) {
		return &r.Floors[len(r.Floors)-1]
	}
	return &r.Floors[floor]
}

func (r *Registry) GetItemDef(defID string) (*ItemDef, bool) {
	idx, ok := r.ItemIDMap[defID]
	if !ok {
		return nil, false
	}
	return &r.Items[idx], true
}

func (r *Registry) GetItemDefByIdx(idx int) *ItemDef {
	if idx < 0 || idx >= len(r.Items) {
		return nil
	}
	return &r.Items[idx]
}

func (r *Registry) GetEnemyDef(defID string) (*ActorDef, bool) {
	idx, ok := r.EnemyIDMap[defID]
	if !ok {
		return nil, false
	}
	return &r.Enemies[idx], true
}

func (r *Registry) GetCompanionDef(defID string) (*ActorDef, bool) {
	idx, ok := r.CompanionIDMap[defID]
	if !ok {
		return nil, false
	}
	return &r.Companions[idx], true
}

func (r *Registry) GetSkillDef(defID string) (*SkillDef, bool) {
	idx, ok := r.SkillIDMap[defID]
	if !ok {
		return nil, false
	}
	return &r.Skills[idx], true
}

func loadSkills(data []byte) ([]SkillDef, map[string]int, error) {
	var raws []rawSkill
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, nil, err
	}
	defs := make([]SkillDef, len(raws))
	idMap := make(map[string]int)
	for i, raw := range raws {
		if _, exists := idMap[raw.ID]; exists {
			return nil, nil, fmt.Errorf("duplicate ID: %s", raw.ID)
		}
		idMap[raw.ID] = i
		defs[i] = SkillDef{
			ID:       raw.ID,
			Name:     raw.Name,
			Desc:     raw.Desc,
			Type:     skillTypeFromString(raw.Type),
			Power:    raw.Power,
			MPCost:   raw.MPCost,
			Range:    raw.Range,
			Element:  component.ElementFromString(raw.Element),
			Status:   raw.Status,
			Duration: raw.Duration,
			ColorHex: raw.Color,
		}
	}
	return defs, idMap, nil
}

func loadRaceBonuses(data []byte) (map[component.Race][6]int, error) {
	var raws []struct {
		Race string `json:"race"`
		Str  int    `json:"str"`
		Wis  int    `json:"wis"`
		Fai  int    `json:"fai"`
		Vit  int    `json:"vit"`
		Agi  int    `json:"agi"`
		Luk  int    `json:"luk"`
	}
	if err := json.Unmarshal(data, &raws); err != nil {
		return nil, err
	}
	res := make(map[component.Race][6]int)
	for _, raw := range raws {
		r := component.RaceFromString(raw.Race)
		res[r] = [6]int{raw.Str, raw.Wis, raw.Fai, raw.Vit, raw.Agi, raw.Luk}
	}
	return res, nil
}
