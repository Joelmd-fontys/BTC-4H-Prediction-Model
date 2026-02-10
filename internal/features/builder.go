package features

import (
	"fmt"
	"math"

	"btc-4h-prediction-model/internal/candles"
)

func BuildFeaturesFromCandles(candleSeries []candles.Candle) ([]FeatureRow, error) {
	if len(candleSeries) < 2 {
		return nil, fmt.Errorf("need at least 2 candles")
	}

	// Returns array aligned to candles (returns[0] is NaN)
	returns := make([]float64, len(candleSeries))
	returns[0] = math.NaN()
	for i := 1; i < len(candleSeries); i++ {
		returns[i] = LogReturn(candleSeries[i].Close, candleSeries[i-1].Close)
	}

	alpha10 := AlphaFromPeriod(10)
	alpha30 := AlphaFromPeriod(30)

	var ema10 float64
	var ema30 float64
	emaInitialized := false

	var rows []FeatureRow

	for i := 0; i < len(candleSeries); i++ {
		c := candleSeries[i]

		row := FeatureRow{
			Exchange:  c.Exchange,
			Symbol:    c.Symbol,
			Timeframe: c.Timeframe,
			Timestamp: c.Timestamp,
		}

		// Range features: available immediately
		rangeHL := (c.High - c.Low) / c.Close
		rangeCO := (c.Close - c.Open) / c.Open
		row.RangeHL = &rangeHL
		row.RangeCO = &rangeCO

		// Volume change: needs previous
		if i >= 1 && candleSeries[i-1].Volume > 0 && c.Volume > 0 {
			volChg := math.Log(c.Volume / candleSeries[i-1].Volume)
			row.VolChg = &volChg
		}

		// Return (1-bar): needs previous
		if i >= 1 {
			ret1 := returns[i]
			row.Ret1 = &ret1
		}

		// Momentum over 6 bars: sum last 6 returns (i-5..i)
		if i >= 6 {
			var sum float64
			for k := i - 5; k <= i; k++ {
				sum += returns[k]
			}
			mom6 := sum
			row.Mom6 = &mom6
		}

		// Volatility over 20 bars
		if i >= 20 {
			vol20 := RollingStd(returns, i, 20)
			row.Vol20 = &vol20
		}

		// EMA features: initialize on first close
		if !emaInitialized {
			ema10 = c.Close
			ema30 = c.Close
			emaInitialized = true
		} else {
			ema10 = EMA(ema10, c.Close, alpha10)
			ema30 = EMA(ema30, c.Close, alpha30)
		}

		ema10Copy := ema10
		ema30Copy := ema30
		row.Ema10 = &ema10Copy
		row.Ema30 = &ema30Copy

		emaSpread := ema10 - ema30
		row.EmaSpread = &emaSpread

		rows = append(rows, row)
	}

	return rows, nil
}
