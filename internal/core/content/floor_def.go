package content

type PoolEntry struct {
	ID     string `json:"id"`
	Weight int    `json:"weight"`
}

type FloorDef struct {
	Floor      int         `json:"floor"`
	MinRooms   int         `json:"min_rooms"`
	MaxRooms   int         `json:"max_rooms"`
	EnemyCount int         `json:"enemy_count"`
	ItemCount  int         `json:"item_count"`
	EnemyPool  []PoolEntry `json:"enemy_pool"`
	ItemPool   []PoolEntry `json:"item_pool"`
}
