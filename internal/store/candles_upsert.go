package store

import (
	"BTC-4H-Prediction-Model/internal/candles"
	"context"
	"database/sql"
	"fmt"
)

func UpsertCandles(
	context context.Context,
	database *sql.DB,
	candlesToUpsert []candles.Candle,
) error {

	if len(candlesToUpsert) == 0 {
		return nil
	}

	const insertOrUpdateQuery = `
INSERT INTO candles (
  exchange,
  symbol,
  timeframe,
  timestamp,
  open,
  high,
  low,
  close,
  volume,
  close_time,
  is_final
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(exchange, symbol, timeframe, timestamp) DO UPDATE SET
  open       = excluded.open,
  high       = excluded.high,
  low        = excluded.low,
  close      = excluded.close,
  volume     = excluded.volume,
  close_time = excluded.close_time,
  is_final   = excluded.is_final;
`

	transaction, beginError := database.BeginTx(context, nil)
	if beginError != nil {
		return beginError
	}
	defer func() { _ = transaction.Rollback() }()

	statement, prepareError := transaction.PrepareContext(context, insertOrUpdateQuery)
	if prepareError != nil {
		return prepareError
	}
	defer statement.Close()

	for _, candle := range candlesToUpsert {
		isFinalInteger := 0
		if candle.IsFinal {
			isFinalInteger = 1
		}

		_, execError := statement.ExecContext(
			context,
			candle.Exchange,
			candle.Symbol,
			candle.Timeframe,
			candle.Timestamp,
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
			candle.CloseTime,
			isFinalInteger,
		)
		if execError != nil {
			return fmt.Errorf(
				"failed to upsert candle (exchange=%s symbol=%s timeframe=%s timestamp=%d): %w",
				candle.Exchange,
				candle.Symbol,
				candle.Timeframe,
				candle.Timestamp,
				execError,
			)
		}
	}

	return transaction.Commit()
}
