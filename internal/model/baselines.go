package model

import "math/rand"

func PredictAlwaysNoTrade(rows []DatasetRow) []Class {
	out := make([]Class, len(rows))
	for i := range out {
		out[i] = ClassNoTrade
	}
	return out
}

func PredictRandomByTrainDistribution(train []DatasetRow, test []DatasetRow, rng *rand.Rand) []Class {
	counts := [3]int{}
	for _, r := range train {
		counts[int(r.Label)]++
	}
	total := counts[0] + counts[1] + counts[2]
	if total == 0 {
		return PredictAlwaysNoTrade(test)
	}

	pUp := float64(counts[0]) / float64(total)
	pDown := float64(counts[1]) / float64(total)
	// pNoTrade is remainder

	out := make([]Class, len(test))
	for i := range test {
		u := rng.Float64()
		if u < pUp {
			out[i] = ClassUp
		} else if u < pUp+pDown {
			out[i] = ClassDown
		} else {
			out[i] = ClassNoTrade
		}
	}
	return out
}
