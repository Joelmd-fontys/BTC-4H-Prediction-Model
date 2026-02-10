package store

import (
	"context"
	"database/sql"
	"fmt"

	"btc-4h-prediction-model/internal/labels"
)

func UpsertLabels(ctx context.Context, db *sql.DB, labelRows []labels.LabelRow) error {
	if len(labelRows) == 0 {
		return nil
	}

	const query = `
INSERT INTO labels (
  exchange, symbol, timeframe, timestamp,
  fwd_ret, label, threshold_b
) VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(exchange, symbol, timeframe, timestamp) DO UPDATE SET
  fwd_ret     = excluded.fwd_ret,
  label       = excluded.label,
  threshold_b = excluded.threshold_b;
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

	for _, row := range labelRows {
		_, execErr := stmt.ExecContext(
			ctx,
			row.Exchange, row.Symbol, row.Timeframe, row.Timestamp,
			row.ForwardReturn, string(row.Label), row.ThresholdB,
		)
		if execErr != nil {
			return fmt.Errorf("upsert label failed timestamp=%d: %w", row.Timestamp, execErr)
		}
	}

	return tx.Commit()
}
