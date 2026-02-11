package backtest

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type equityPoint struct {
	Timestamp int64
	Equity    float64
	Traded    int
	Side      string
	TradeRet  float64
}

type PredictionWithReturn struct {
	Timestamp int64

	PUp   float64
	PDown float64

	ActualLabel    string
	PredictedLabel string

	// labels.fwd_ret is a log return from t -> t+1
	ForwardLogReturn sql.NullFloat64
}

type PaperConfig struct {
	Exchange  string
	Symbol    string
	Timeframe string
	ModelName string

	Threshold float64

	// Costs as decimals, e.g. 0.0004 = 4 bps
	FeePerSide float64
	Slippage   float64

	// Where to write equity CSV (optional; empty disables)
	EquityCSVPath string
}

type PaperResult struct {
	Threshold float64

	Predictions int
	Trades      int
	Coverage    float64

	TotalReturn float64 // ending_equity - 1
	Sharpe      float64
	MaxDrawdown float64

	EndEquity float64
}

func LoadPredictionsWithForwardReturn(
	ctx context.Context,
	db *sql.DB,
	exchange string,
	symbol string,
	timeframe string,
	modelName string,
) ([]PredictionWithReturn, error) {

	query := `
SELECT
  p.timestamp,
  p.p_up,
  p.p_down,
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

	var out []PredictionWithReturn
	for rows.Next() {
		var r PredictionWithReturn
		if err := rows.Scan(
			&r.Timestamp,
			&r.PUp,
			&r.PDown,
			&r.PredictedLabel,
			&r.ActualLabel,
			&r.ForwardLogReturn,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func RunPaperBacktest(rows []PredictionWithReturn, cfg PaperConfig) (PaperResult, error) {
	if cfg.Threshold <= 0 || cfg.Threshold >= 1 {
		return PaperResult{}, fmt.Errorf("threshold must be in (0,1)")
	}
	if cfg.FeePerSide < 0 || cfg.Slippage < 0 {
		return PaperResult{}, fmt.Errorf("fee/slippage must be >= 0")
	}

	// Ensure sorted
	sort.Slice(rows, func(i, j int) bool { return rows[i].Timestamp < rows[j].Timestamp })

	equity := 1.0
	peak := 1.0
	maxDD := 0.0

	strategyReturns := make([]float64, 0, len(rows))
	trades := 0

	// Costs: apply on entry + exit => 2*fee, plus slippage (we treat as round-trip for simplicity)
	roundTripCost := 2.0*cfg.FeePerSide + cfg.Slippage

	var curve []equityPoint
	curve = append(curve, equityPoint{Timestamp: rows[0].Timestamp, Equity: equity, Traded: 0, Side: "NONE", TradeRet: 0})

	for _, r := range rows {
		// Confidence uses direction probabilities
		conf := r.PUp
		side := "UP"
		if r.PDown > conf {
			conf = r.PDown
			side = "DOWN"
		}

		traded := 0
		tradeRet := 0.0

		// default: no position, flat return
		periodReturn := 0.0

		// Only trade if confident enough AND we have realized forward return
		if conf >= cfg.Threshold && r.ForwardLogReturn.Valid {
			traded = 1
			trades++

			// Convert log return to simple return for this 4H period
			fwdSimple := math.Exp(r.ForwardLogReturn.Float64) - 1.0

			// Long on UP, short on DOWN (profit is -return for short)
			dir := 1.0
			if side == "DOWN" {
				dir = -1.0
			}

			tradeRet = dir*fwdSimple - roundTripCost
			periodReturn = tradeRet
		}

		// Apply return to equity (cap at 0 to avoid negative equity due to costs on tiny equity)
		equity = equity * (1.0 + periodReturn)
		if equity < 0 {
			equity = 0
		}

		// Drawdown
		if equity > peak {
			peak = equity
		}
		dd := 0.0
		if peak > 0 {
			dd = (peak - equity) / peak
		}
		if dd > maxDD {
			maxDD = dd
		}

		strategyReturns = append(strategyReturns, periodReturn)
		curve = append(curve, equityPoint{
			Timestamp: r.Timestamp,
			Equity:    equity,
			Traded:    traded,
			Side:      side,
			TradeRet:  tradeRet,
		})
	}

	// Sharpe annualized on 4H periods (~2190 per year)
	sharpe := annualizedSharpe(strategyReturns, 2190.0)

	result := PaperResult{
		Threshold: cfg.Threshold,

		Predictions: len(rows),
		Trades:      trades,
		Coverage:    float64(trades) / float64(len(rows)),

		EndEquity:   equity,
		TotalReturn: equity - 1.0,
		MaxDrawdown: maxDD,
		Sharpe:      sharpe,
	}

	if cfg.EquityCSVPath != "" {
		if err := writeEquityCSV(cfg.EquityCSVPath, curve); err != nil {
			return PaperResult{}, err
		}
	}

	return result, nil
}

func annualizedSharpe(returns []float64, periodsPerYear float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	var s2 float64
	for _, r := range returns {
		d := r - mean
		s2 += d * d
	}
	variance := s2 / float64(len(returns))
	std := math.Sqrt(variance)
	if std == 0 {
		return 0
	}
	return (mean / std) * math.Sqrt(periodsPerYear)
}

func writeEquityCSV(path string, points []equityPoint) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"timestamp", "equity", "traded", "side", "trade_return"})

	for _, p := range points {
		_ = w.Write([]string{
			strconv.FormatInt(p.Timestamp, 10),
			fmt.Sprintf("%.10f", p.Equity),
			strconv.Itoa(p.Traded),
			p.Side,
			fmt.Sprintf("%.10f", p.TradeRet),
		})
	}

	return w.Error()
}
