package model

type PredictionRow struct {
	Exchange  string
	Symbol    string
	Timeframe string
	Timestamp int64
	ModelName string

	PUp      float64
	PDown    float64
	PNoTrade float64

	Predicted Class
	Actual    Class
}
