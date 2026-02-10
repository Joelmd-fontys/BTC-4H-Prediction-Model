package model

type Class int

const (
	ClassUp Class = iota
	ClassDown
	ClassNoTrade
)

func (c Class) String() string {
	switch c {
	case ClassUp:
		return "UP"
	case ClassDown:
		return "DOWN"
	default:
		return "NO_TRADE"
	}
}

func ParseLabel(label string) Class {
	switch label {
	case "UP":
		return ClassUp
	case "DOWN":
		return ClassDown
	default:
		return ClassNoTrade
	}
}

type DatasetRow struct {
	Exchange  string
	Symbol    string
	Timeframe string
	Timestamp int64

	// Features used for training
	Ret1      float64
	Vol20     float64
	Mom6      float64
	EmaSpread float64
	RangeHL   float64
	RangeCO   float64
	VolChg    float64

	Label Class
}
