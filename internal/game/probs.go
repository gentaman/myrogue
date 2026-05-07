package game

import (
	"math/rand"
	"slices"
	"sort"
)

func weightedChoice[T any](rng *rand.Rand, items []T, weights []int) int {
	if len(weights) <= 0 {
		return -1
	}

	minW := slices.Min(weights)

	if minW < 0 {
		return -1
	}

	// 1. 累積和の配列を作成
	sums := make([]int, len(weights))
	runningSum := 0
	for i, w := range weights {
		runningSum += w
		sums[i] = runningSum
	}

	// 2. 0 〜 総和 の範囲で乱数を生成
	r := rng.Intn(runningSum)

	// 3. 二分探索でどの範囲に含まれるか特定
	idx := sort.Search(len(sums), func(i int) bool {
		return sums[i] > r
	})

	return idx
}
