package store

import (
	"context"
	"database/sql"
	"fmt"

	"btc-4h-prediction-model/internal/model"
)

func UpsertPredictions(ctx context.Context, db *sql.DB, rows []model.PredictionRow) error {
	if len(rows) == 0 {
		return nil
	}

	const query = `
INSERT INTO predictions (
  exchange, symbol, timeframe, timestamp,
  model_name,
  p_up, p_down, p_no_trade,
  predicted_label, actual_label
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(exchange, symbol, timeframe, timestamp, model_name) DO UPDATE SET
  p_up = excluded.p_up,
  p_down = excluded.p_down,
  p_no_trade = excluded.p_no_trade,
  predicted_label = excluded.predicted_label,
  actual_label = excluded.actual_label;
`

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, row := range rows {
		_, execErr := stmt.ExecContext(
			ctx,
			row.Exchange, row.Symbol, row.Timeframe, row.Timestamp,
			row.ModelName,
			row.PUp, row.PDown, row.PNoTrade,
			row.Predicted.String(),
			row.Actual.String(),
		)
		if execErr != nil {
			return fmt.Errorf("upsert prediction failed timestamp=%d: %w", row.Timestamp, execErr)
		}
	}

	return tx.Commit()
}
