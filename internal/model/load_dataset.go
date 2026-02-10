package model

import (
	"context"
	"database/sql"
	"fmt"
)

func LoadDatasetOrdered(
	ctx context.Context,
	db *sql.DB,
	exchange string,
	symbol string,
	timeframe string,
) ([]DatasetRow, error) {

	rows, err := db.QueryContext(ctx, `
SELECT
  exchange, symbol, timeframe, timestamp,
  ret_1, vol_20, mom_6, ema_spread, range_hl, range_co, vol_chg,
  label
FROM dataset
WHERE exchange = ? AND symbol = ? AND timeframe = ?
ORDER BY timestamp ASC;
`, exchange, symbol, timeframe)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []DatasetRow
	for rows.Next() {
		var row DatasetRow
		var label string

		if err := rows.Scan(
			&row.Exchange, &row.Symbol, &row.Timeframe, &row.Timestamp,
			&row.Ret1, &row.Vol20, &row.Mom6, &row.EmaSpread, &row.RangeHL, &row.RangeCO, &row.VolChg,
			&label,
		); err != nil {
			return nil, err
		}

		row.Label = ParseLabel(label)
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("dataset empty for %s %s %s", exchange, symbol, timeframe)
	}
	return result, nil
}
