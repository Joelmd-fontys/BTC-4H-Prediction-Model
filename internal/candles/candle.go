package candles

type Candle struct {
	Exchange  string
	Symbol    string
	Timeframe string

	Timestamp int64 // openTime in ms (UTC)
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64

	CloseTime int64 // optional, 0 if unknown
	IsFinal   bool
}
