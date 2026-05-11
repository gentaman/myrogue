package world

type TileType int

const (
	Wall TileType = iota
	Floor
	Stairs
	StairsDown
	StairsUp
)
