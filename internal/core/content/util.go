package content

import (
	"math/rand"
	"sort"
)

func WeightedChoice(rng *rand.Rand, weights []int) int {
	if len(weights) == 0 {
		return -1
	}
	sums := make([]int, len(weights))
	runningSum := 0
	for i, w := range weights {
		runningSum += w
		sums[i] = runningSum
	}
	if runningSum <= 0 {
		return -1
	}
	r := rng.Intn(runningSum)
	idx := sort.Search(len(sums), func(i int) bool {
		return sums[i] > r
	})
	return idx
}
