package features

type FeatureRow struct {
	Exchange  string
	Symbol    string
	Timeframe string
	Timestamp int64

	Ret1      *float64
	Vol20     *float64
	Mom6      *float64
	Ema10     *float64
	Ema30     *float64
	EmaSpread *float64
	RangeHL   *float64
	RangeCO   *float64
	VolChg    *float64
}
