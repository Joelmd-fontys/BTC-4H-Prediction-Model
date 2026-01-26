package store

import (
	"BTC-4H-Prediction-Model/internal/candles"
	"context"
	"database/sql"
)

func InsertOneCandle(ctx context.Context, db *sql.DB, c candles.Candle) error {
	const q = `
INSERT INTO candles (
  exchange, symbol, timeframe, timestamp,
  open, high, low, close, volume,
  close_time, is_final
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	isFinal := 0
	if c.IsFinal {
		isFinal = 1
	}

	_, err := db.ExecContext(
		ctx,
		q,
		c.Exchange, c.Symbol, c.Timeframe, c.Timestamp,
		c.Open, c.High, c.Low, c.Close, c.Volume,
		c.CloseTime, isFinal,
	)
	return err
}
