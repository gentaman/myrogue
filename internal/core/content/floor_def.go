package content

type PoolEntry struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
}

type ConditionalItem struct {
	ID        string `json:"id"`
	Condition string `json:"condition"`
}

type DropConditionChecker interface {
	CheckCondition(condition, itemID string) bool
}

type FloorDef struct {
	Floor            int               `json:"floor"`
	MinRooms         int               `json:"min_rooms"`
	MaxRooms         int               `json:"max_rooms"`
	EnemyCount       int               `json:"enemy_count"`
	ItemCount        int               `json:"item_count"`
	EnemyPool        []PoolEntry       `json:"enemy_pool"`
	ItemPool         []PoolEntry       `json:"item_pool"`
	ConditionalItems []ConditionalItem `json:"conditional_items"`
}
