package exchange

import (
	"BTC-4H-Prediction-Model/internal/candles"
	"encoding/json"
	"fmt"
	"strconv"
)

func ParseKlinesToCandles(body []byte, exchangeName, symbol, timeframe string) ([]candles.Candle, error) {
	var rows [][]interface{}
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}

	var result []candles.Candle

	for _, r := range rows {
		if len(r) < 7 {
			return nil, fmt.Errorf("unexpected kline length: %d", len(r))
		}

		// IDEA: make a function that handles the conversion -> factory, modular
		openTime := int64(r[0].(float64))
		open, err := strconv.ParseFloat(r[1].(string), 64)
		if err != nil {
			return nil, err
		}
		high, err := strconv.ParseFloat(r[2].(string), 64)
		if err != nil {
			return nil, err
		}
		low, err := strconv.ParseFloat(r[3].(string), 64)
		if err != nil {
			return nil, err
		}
		closeP, err := strconv.ParseFloat(r[4].(string), 64)
		if err != nil {
			return nil, err
		}
		vol, err := strconv.ParseFloat(r[5].(string), 64)
		if err != nil {
			return nil, err
		}
		closeTime := int64(r[6].(float64))

		c := candles.Candle{
			Exchange:  exchangeName,
			Symbol:    symbol,
			Timeframe: timeframe,
			Timestamp: openTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closeP,
			Volume:    vol,
			CloseTime: closeTime,
			IsFinal:   true,
		}

		result = append(result, c)
	}

	return result, nil
}

/* temporary
func ParseFirstKlineToCandle(body []byte, exchangeName, symbol, timeframe string) (candles.Candle, error) {
	var rows [][]interface{}
	if err := json.Unmarshal(body, &rows); err != nil {
		return candles.Candle{}, err
	}
	if len(rows) == 0 {
		return candles.Candle{}, fmt.Errorf("no klines returned")
	}

	r := rows[0]
	if len(r) < 7 {
		return candles.Candle{}, fmt.Errorf("unexpected kline length: %d", len(r))
	}

	// , _ -> errors not handled "unstable"
	openTime := int64(r[0].(float64))
	openPrice, _ := strconv.ParseFloat(r[1].(string), 64)
	highPrice, _ := strconv.ParseFloat(r[2].(string), 64)
	lowPrice, _ := strconv.ParseFloat(r[3].(string), 64)
	closePrice, _ := strconv.ParseFloat(r[4].(string), 64)
	vol, _ := strconv.ParseFloat(r[5].(string), 64)
	closeTime := int64(r[6].(float64))

	return candles.Candle{
		Exchange:  exchangeName,
		Symbol:    symbol,
		Timeframe: timeframe,

		Timestamp: openTime,
		Open:      openPrice,
		High:      highPrice,
		Low:       lowPrice,
		Close:     closePrice,
		Volume:    vol,
		CloseTime: closeTime,
		IsFinal:   true, // v0.1: store only closed candles
	}, nil
}
*/
