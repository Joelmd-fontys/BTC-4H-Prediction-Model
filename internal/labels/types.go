package labels

type Label string

const (
	LabelUp      Label = "UP"
	LabelDown    Label = "DOWN"
	LabelNoTrade Label = "NO_TRADE"
)

type LabelRow struct {
	Exchange  string
	Symbol    string
	Timeframe string
	Timestamp int64

	ForwardReturn float64
	ThresholdB    float64
	Label         Label
}
