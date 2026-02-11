package cli

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/store"
)

type confidenceRow struct {
	Timestamp int64

	PUp      float64
	PDown    float64
	PNoTrade float64

	PredictedLabel string
	ActualLabel    string

	ForwardReturn sql.NullFloat64
}

type confidenceStats struct {
	Threshold float64

	TotalPredictions int
	Trades           int
	Coverage         float64

	DirectionalPrecision float64 // among trades only
	DirectionalRecall    float64 // among actual UP/DOWN only (optional-ish)

	AverageForwardReturn float64 // among trades (if joined)
}

var confidenceSymbol string
var confidenceTimeframe string
var confidenceModelName string
var confidenceMinThreshold float64
var confidenceMaxThreshold float64
var confidenceStep float64

var confidenceCommand = &cobra.Command{
	Use:   "confidence",
	Short: "Analyze prediction confidence thresholds (coverage/precision/returns) from stored predictions",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		db, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer db.Close()

		if confidenceStep <= 0 {
			return fmt.Errorf("--step must be > 0")
		}
		if confidenceMinThreshold <= 0 || confidenceMinThreshold >= 1 {
			return fmt.Errorf("--min must be in (0,1)")
		}
		if confidenceMaxThreshold <= 0 || confidenceMaxThreshold > 1 {
			return fmt.Errorf("--max must be in (0,1]")
		}
		if confidenceMinThreshold > confidenceMaxThreshold {
			return fmt.Errorf("--min must be <= --max")
		}

		rows, err := loadConfidenceRows(ctx, db, "binance", confidenceSymbol, confidenceTimeframe, confidenceModelName)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("no predictions found for model=%s %s %s", confidenceModelName, confidenceSymbol, confidenceTimeframe)
		}

		thresholds := buildThresholds(confidenceMinThreshold, confidenceMaxThreshold, confidenceStep)
		stats := computeConfidenceStats(rows, thresholds)

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", confidenceSymbol)
		fmt.Println("timeframe:", confidenceTimeframe)
		fmt.Println("model:", confidenceModelName)
		fmt.Println("predictions:", len(rows))
		fmt.Println()

		// Print a readable table
		fmt.Printf("%-8s %-10s %-10s %-12s %-14s %-14s %-14s\n",
			"thr",
			"trades",
			"coverage",
			"dir_prec",
			"dir_recall",
			"avg_fwd_ret",
			"notes",
		)

		for _, s := range stats {
			note := ""
			if s.Trades < 50 {
				note = "low-n"
			}
			fmt.Printf("%-8.2f %-10d %-10.3f %-12.3f %-14.3f %-14.5f %-14s\n",
				s.Threshold,
				s.Trades,
				s.Coverage,
				s.DirectionalPrecision,
				s.DirectionalRecall,
				s.AverageForwardReturn,
				note,
			)
		}

		fmt.Println()
		fmt.Println("Interpretation tips:")
		fmt.Println("- coverage = trades / predictions (higher threshold => fewer trades)")
		fmt.Println("- dir_prec = P(actual matches direction | traded)")
		fmt.Println("- avg_fwd_ret uses label forward return (simple 4H horizon proxy, not full trading sim)")
		fmt.Println("- Focus on thresholds where trades are not 'low-n' (e.g. >= 50).")

		return nil
	},
}

func init() {
	confidenceCommand.Flags().StringVar(&confidenceSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	confidenceCommand.Flags().StringVar(&confidenceTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")
	confidenceCommand.Flags().StringVar(&confidenceModelName, "model", "logreg_softmax", "Model name as stored in predictions.model_name")

	confidenceCommand.Flags().Float64Var(&confidenceMinThreshold, "min", 0.40, "Minimum confidence threshold")
	confidenceCommand.Flags().Float64Var(&confidenceMaxThreshold, "max", 0.80, "Maximum confidence threshold")
	confidenceCommand.Flags().Float64Var(&confidenceStep, "step", 0.05, "Step size for thresholds")
}

func loadConfidenceRows(
	ctx context.Context,
	db *sql.DB,
	exchange string,
	symbol string,
	timeframe string,
	modelName string,
) ([]confidenceRow, error) {

	// Join labels to get forward return (fwd_ret). This is safe because it’s the realized next-horizon return,
	// and we’re analyzing after the fact.
	query := `
SELECT
  p.timestamp,
  p.p_up,
  p.p_down,
  p.p_no_trade,
  p.predicted_label,
  p.actual_label,
  l.fwd_ret
FROM predictions p
LEFT JOIN labels l
  ON p.exchange = l.exchange
 AND p.symbol = l.symbol
 AND p.timeframe = l.timeframe
 AND p.timestamp = l.timestamp
WHERE p.exchange = ? AND p.symbol = ? AND p.timeframe = ? AND p.model_name = ?
ORDER BY p.timestamp ASC;
`

	rows, err := db.QueryContext(ctx, query, exchange, symbol, timeframe, modelName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []confidenceRow
	for rows.Next() {
		var r confidenceRow
		if err := rows.Scan(
			&r.Timestamp,
			&r.PUp,
			&r.PDown,
			&r.PNoTrade,
			&r.PredictedLabel,
			&r.ActualLabel,
			&r.ForwardReturn,
		); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func buildThresholds(min float64, max float64, step float64) []float64 {
	var thresholds []float64
	// Use a loop that’s robust to float stepping issues
	for t := min; t <= max+1e-9; t += step {
		thresholds = append(thresholds, math.Round(t*100.0)/100.0) // round to 2 decimals for display
	}
	// Ensure unique + sorted
	sort.Float64s(thresholds)
	thresholds = uniqueFloat64(thresholds)
	return thresholds
}

func uniqueFloat64(values []float64) []float64 {
	if len(values) == 0 {
		return values
	}
	out := []float64{values[0]}
	for i := 1; i < len(values); i++ {
		if math.Abs(values[i]-values[i-1]) > 1e-9 {
			out = append(out, values[i])
		}
	}
	return out
}

func computeConfidenceStats(rows []confidenceRow, thresholds []float64) []confidenceStats {
	total := len(rows)

	// Count actual direction events (UP/DOWN) for recall denominator
	actualDirectionalCount := 0
	for _, r := range rows {
		if r.ActualLabel == "UP" || r.ActualLabel == "DOWN" {
			actualDirectionalCount++
		}
	}

	var result []confidenceStats

	for _, threshold := range thresholds {
		trades := 0
		correctDirectional := 0

		// Directional recall numerator: correct direction on actual UP/DOWN events (but only when we traded)
		correctDirectionalOnActualDirectional := 0

		sumForwardReturn := 0.0
		forwardReturnCount := 0

		for _, r := range rows {
			// Define "confidence" as max(p_up, p_down), since NO_TRADE is "do nothing"
			confidence := r.PUp
			predictedDirection := "UP"
			if r.PDown > confidence {
				confidence = r.PDown
				predictedDirection = "DOWN"
			}

			// Trade only if confident enough AND predicted direction isn’t NO_TRADE
			if confidence < threshold {
				continue
			}

			trades++

			// Directional correctness: predicted UP matches actual UP (same for DOWN)
			if predictedDirection == r.ActualLabel {
				correctDirectional++
			}
			if (r.ActualLabel == "UP" || r.ActualLabel == "DOWN") && predictedDirection == r.ActualLabel {
				correctDirectionalOnActualDirectional++
			}

			// Avg forward return proxy:
			// If we predicted DOWN, we treat return as "short": profit is -fwd_ret.
			// If UP: profit is +fwd_ret.
			if r.ForwardReturn.Valid {
				tradeReturn := r.ForwardReturn.Float64
				if predictedDirection == "DOWN" {
					tradeReturn = -tradeReturn
				}
				sumForwardReturn += tradeReturn
				forwardReturnCount++
			}
		}

		coverage := 0.0
		if total > 0 {
			coverage = float64(trades) / float64(total)
		}

		directionalPrecision := 0.0
		if trades > 0 {
			directionalPrecision = float64(correctDirectional) / float64(trades)
		}

		directionalRecall := 0.0
		if actualDirectionalCount > 0 {
			directionalRecall = float64(correctDirectionalOnActualDirectional) / float64(actualDirectionalCount)
		}

		avgForwardReturn := 0.0
		if forwardReturnCount > 0 {
			avgForwardReturn = sumForwardReturn / float64(forwardReturnCount)
		}

		result = append(result, confidenceStats{
			Threshold: threshold,

			TotalPredictions: total,
			Trades:           trades,
			Coverage:         coverage,

			DirectionalPrecision: directionalPrecision,
			DirectionalRecall:    directionalRecall,

			AverageForwardReturn: avgForwardReturn,
		})
	}

	return result
}
