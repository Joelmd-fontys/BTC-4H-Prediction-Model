package store

import (
	"context"
	"database/sql"
	"fmt"

	"btc-4h-prediction-model/internal/features"
)

func UpsertFeatures(ctx context.Context, db *sql.DB, rows []features.FeatureRow) error {
	if len(rows) == 0 {
		return nil
	}

	const query = `
INSERT INTO features (
  exchange, symbol, timeframe, timestamp,
  ret_1, vol_20, mom_6, ema_10, ema_30, ema_spread,
  range_hl, range_co, vol_chg
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(exchange, symbol, timeframe, timestamp) DO UPDATE SET
  ret_1      = excluded.ret_1,
  vol_20     = excluded.vol_20,
  mom_6      = excluded.mom_6,
  ema_10     = excluded.ema_10,
  ema_30     = excluded.ema_30,
  ema_spread = excluded.ema_spread,
  range_hl   = excluded.range_hl,
  range_co   = excluded.range_co,
  vol_chg    = excluded.vol_chg;
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
			row.Ret1, row.Vol20, row.Mom6, row.Ema10, row.Ema30, row.EmaSpread,
			row.RangeHL, row.RangeCO, row.VolChg,
		)
		if execErr != nil {
			return fmt.Errorf("upsert features failed timestamp=%d: %w", row.Timestamp, execErr)
		}
	}

	return tx.Commit()
}
