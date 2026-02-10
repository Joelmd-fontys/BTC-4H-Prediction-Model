package features

import "math"

func LogReturn(currentClose float64, previousClose float64) float64 {
	return math.Log(currentClose / previousClose)
}

func AlphaFromPeriod(period int) float64 {
	return 2.0 / (float64(period) + 1.0)
}

func EMA(previousEMA float64, price float64, alpha float64) float64 {
	return alpha*price + (1.0-alpha)*previousEMA
}

func RollingStd(values []float64, i int, window int) float64 {
	start := i - window + 1

	var sum float64
	for j := start; j <= i; j++ {
		sum += values[j]
	}
	mean := sum / float64(window)

	var varianceSum float64
	for j := start; j <= i; j++ {
		diff := values[j] - mean
		varianceSum += diff * diff
	}
	return math.Sqrt(varianceSum / float64(window))
}
