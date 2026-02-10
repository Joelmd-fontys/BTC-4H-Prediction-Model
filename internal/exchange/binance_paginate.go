package exchange

import (
	"BTC-4H-Prediction-Model/internal/candles"
	"context"
	"fmt"
)

const (
	binanceKlinesMaxLimit = 1000
)

func IntervalToMillis(interval string) (int64, error) {
	// v0.1 only needs 4h, keep it strict to avoid mistakes
	if interval == "4h" {
		return 4 * 60 * 60 * 1000, nil
	}
	return 0, fmt.Errorf("unsupported interval: %s", interval)
}

func (client BinanceClient) FetchKlinesPaginated(
	context context.Context,
	symbol string,
	interval string,
	startTimeMillis int64,
	endTimeMillis int64,
) ([]candles.Candle, error) {

	intervalMillis, err := IntervalToMillis(interval)
	if err != nil {
		return nil, err
	}

	currentStartTimeMillis := startTimeMillis
	var allCandles []candles.Candle

	for {
		requestURL := BuildKlinesURL(symbol, interval, binanceKlinesMaxLimit, currentStartTimeMillis, endTimeMillis)

		body, err := client.Get(context, requestURL)
		if err != nil {
			return nil, err
		}

		pageCandles, err := ParseKlinesToCandles(body, "binance", symbol, interval)
		if err != nil {
			return nil, err
		}

		// Stop if no data returned
		if len(pageCandles) == 0 {
			break
		}

		allCandles = append(allCandles, pageCandles...)

		// Move start forward to avoid duplicates
		lastCandle := pageCandles[len(pageCandles)-1]
		nextStartTimeMillis := lastCandle.Timestamp + intervalMillis

		// Safety: avoid infinite loop
		if nextStartTimeMillis <= currentStartTimeMillis {
			break
		}
		currentStartTimeMillis = nextStartTimeMillis

		// If less than max page size, we’re likely done
		if len(pageCandles) < binanceKlinesMaxLimit {
			break
		}

		// If endTime is set and we’ve passed it, stop
		if endTimeMillis > 0 && currentStartTimeMillis > endTimeMillis {
			break
		}
	}

	return allCandles, nil
}
