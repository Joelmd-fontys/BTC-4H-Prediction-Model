package labels

import (
	"fmt"
	"math"

	"btc-4h-prediction-model/internal/candles"
)

func BuildForwardReturnLabels(
	candleSeries []candles.Candle,
	thresholdB float64,
) ([]LabelRow, error) {

	if thresholdB <= 0 {
		return nil, fmt.Errorf("thresholdB must be > 0")
	}
	if len(candleSeries) < 2 {
		return nil, fmt.Errorf("need at least 2 candles to build labels")
	}

	var rows []LabelRow

	// For each t, use close_{t+1} and close_{t}
	// Last candle has no label (no future candle), so stop at len-2.
	for i := 0; i < len(candleSeries)-1; i++ {
		current := candleSeries[i]
		next := candleSeries[i+1]

		if current.Close <= 0 || next.Close <= 0 {
			return nil, fmt.Errorf("non-positive close found at timestamp=%d", current.Timestamp)
		}

		forwardReturn := math.Log(next.Close / current.Close)

		var label Label
		if forwardReturn > thresholdB {
			label = LabelUp
		} else if forwardReturn < -thresholdB {
			label = LabelDown
		} else {
			label = LabelNoTrade
		}

		rows = append(rows, LabelRow{
			Exchange:      current.Exchange,
			Symbol:        current.Symbol,
			Timeframe:     current.Timeframe,
			Timestamp:     current.Timestamp,
			ForwardReturn: forwardReturn,
			ThresholdB:    thresholdB,
			Label:         label,
		})
	}

	return rows, nil
}
