package rules

import "math/rand"

type RNG interface {
	Intn(n int) int
	Float64() float64
}

type StdRNG struct {
	src *rand.Rand
}

func NewRNG(seed int64) *StdRNG {
	return &StdRNG{src: rand.New(rand.NewSource(seed))}
}

func (r *StdRNG) Intn(n int) int {
	return r.src.Intn(n)
}

func (r *StdRNG) Float64() float64 {
	return r.src.Float64()
}

func (r *StdRNG) Seed() int64 {
	return 0
}
