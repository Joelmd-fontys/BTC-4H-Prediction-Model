package store

import (
	"context"
	"database/sql"

	"btc-4h-prediction-model/internal/candles"
)

func LoadCandlesOrdered(ctx context.Context, db *sql.DB, exchange string, symbol string, timeframe string) ([]candles.Candle, error) {
	rows, err := db.QueryContext(ctx, `
SELECT exchange, symbol, timeframe, timestamp,
       open, high, low, close, volume,
       close_time, is_final
FROM candles
WHERE exchange=? AND symbol=? AND timeframe=?
ORDER BY timestamp ASC;
`, exchange, symbol, timeframe)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []candles.Candle
	for rows.Next() {
		var candle candles.Candle
		var isFinalInt int

		if err := rows.Scan(
			&candle.Exchange, &candle.Symbol, &candle.Timeframe, &candle.Timestamp,
			&candle.Open, &candle.High, &candle.Low, &candle.Close, &candle.Volume,
			&candle.CloseTime, &isFinalInt,
		); err != nil {
			return nil, err
		}
		candle.IsFinal = (isFinalInt == 1)
		result = append(result, candle)
	}

	return result, rows.Err()
}
