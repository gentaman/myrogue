package component

import "math"

const (
	MaxStr = 100
	MaxWis = 100
	MaxFai = 100
	MaxVit = 100
	MaxAgi = 100
	MaxLuk = 100
)

type Stats struct {
	HP, MaxHP int
	MP, MaxMP int
	Level     int
	XP        int
	XPToNext  int
	Str       int
	Wis       int
	Fai       int
	Vit       int
	Agi       int
	Luk       int
}

func (s *Stats) CalcXPToNext() {
	s.XPToNext = int(10 * math.Pow(float64(s.Level), 1.5))
}

type CombatStats struct {
	PhysicalAttack  int
	PhysicalDefense int
	MagicalAttack   int
	MagicalDefense  int
	BaseAccuracy    int
}
