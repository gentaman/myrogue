package component

type EquipSlot int

const (
	SlotNone EquipSlot = iota
	SlotWeapon
	SlotShield
	SlotArmor
)

type ItemEntry struct {
	DefID         string
	Count         int
	Durability    int
	Equipped      bool
	ObtainedSeed  int64
	ObtainedFloor int
}

type Inventory struct {
	Items          []ItemEntry
	MaxCarryWeight int
}
