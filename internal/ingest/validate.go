package ingest

import (
	"context"
	"database/sql"
	"fmt"
)

type Gap struct {
	PreviousTS int64
	ExpectedTS int64
	ActualTS   int64
	Missing    int64 // number of missing intervals
}

type ValidationResult struct {
	Exchange  string
	Symbol    string
	Timeframe string

	Count int64

	FirstTS int64
	LastTS  int64

	Gaps []Gap
}

// ValidateCandleContinuity checks that ts is strictly increasing and spaced by expectedIntervalMillis.
// It reports gaps (missing candles). Duplicates are prevented by PK, but this still detects query/order issues.
func ValidateCandleContinuity(
	ctx context.Context,
	db *sql.DB,
	exchange string,
	symbol string,
	timeframe string,
	expectedIntervalMillis int64,
) (ValidationResult, error) {

	rows, err := db.QueryContext(ctx, `
SELECT timestamp 
FROM candles
WHERE exchange = ? AND symbol = ? AND timeframe = ?
ORDER BY timestamp ASC;
`, exchange, symbol, timeframe)
	if err != nil {
		return ValidationResult{}, err
	}
	defer rows.Close()

	var result ValidationResult
	result.Exchange = exchange
	result.Symbol = symbol
	result.Timeframe = timeframe

	var previousTS int64
	havePrevious := false

	for rows.Next() {
		var currentTS int64
		if err := rows.Scan(&currentTS); err != nil {
			return ValidationResult{}, err
		}

		result.Count++

		if result.Count == 1 {
			result.FirstTS = currentTS
		}
		result.LastTS = currentTS

		if havePrevious {
			// Out-of-order / duplicate guard (shouldn't happen with ORDER BY + PK, but validate anyway)
			if currentTS <= previousTS {
				return ValidationResult{}, fmt.Errorf("non-increasing ts detected: prev=%d current=%d", previousTS, currentTS)
			}

			expectedTS := previousTS + expectedIntervalMillis
			if currentTS != expectedTS {
				// Compute how many intervals are missing (could be >1 if large gap)
				delta := currentTS - previousTS
				if delta > expectedIntervalMillis && delta%expectedIntervalMillis == 0 {
					missingIntervals := (delta / expectedIntervalMillis) - 1
					result.Gaps = append(result.Gaps, Gap{
						PreviousTS: previousTS,
						ExpectedTS: expectedTS,
						ActualTS:   currentTS,
						Missing:    missingIntervals,
					})
				} else {
					// Not aligned to interval boundary (weird data)
					result.Gaps = append(result.Gaps, Gap{
						PreviousTS: previousTS,
						ExpectedTS: expectedTS,
						ActualTS:   currentTS,
						Missing:    -1,
					})
				}
			}
		}

		previousTS = currentTS
		havePrevious = true
	}

	if err := rows.Err(); err != nil {
		return ValidationResult{}, err
	}

	return result, nil
}
