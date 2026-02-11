package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/backtest"
	"btc-4h-prediction-model/internal/store"
)

var paperSymbol string
var paperTimeframe string
var paperModelName string
var paperThresholds string
var paperFee float64
var paperSlippage float64
var paperOutDir string

var paperCommand = &cobra.Command{
	Use:   "paper",
	Short: "Paper trading simulator (4H horizon) using stored out-of-sample predictions",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		db, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer db.Close()

		thresholdList, err := parseThresholds(paperThresholds)
		if err != nil {
			return err
		}

		rows, err := backtest.LoadPredictionsWithForwardReturn(ctx, db, "binance", paperSymbol, paperTimeframe, paperModelName)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return fmt.Errorf("no predictions found for %s %s model=%s", paperSymbol, paperTimeframe, paperModelName)
		}

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", paperSymbol)
		fmt.Println("timeframe:", paperTimeframe)
		fmt.Println("model:", paperModelName)
		fmt.Println("predictions:", len(rows))
		fmt.Println("feePerSide:", paperFee, "slippage:", paperSlippage)
		fmt.Println()

		fmt.Printf("%-8s %-10s %-10s %-12s %-12s %-12s %-12s\n",
			"thr", "trades", "coverage", "total_ret", "sharpe", "max_dd", "csv",
		)

		for _, thr := range thresholdList {
			csvPath := ""
			if paperOutDir != "" {
				csvPath = filepath.Join(
					paperOutDir,
					fmt.Sprintf("equity_%s_%s_thr%.2f.csv", strings.ToLower(paperSymbol), strings.ToLower(paperTimeframe), thr),
				)
			}

			result, err := backtest.RunPaperBacktest(rows, backtest.PaperConfig{
				Exchange:  "binance",
				Symbol:    paperSymbol,
				Timeframe: paperTimeframe,
				ModelName: paperModelName,

				Threshold:  thr,
				FeePerSide: paperFee,
				Slippage:   paperSlippage,

				EquityCSVPath: csvPath,
			})
			if err != nil {
				return err
			}

			fmt.Printf("%-8.2f %-10d %-10.3f %-12.4f %-12.4f %-12.4f %-12s\n",
				result.Threshold,
				result.Trades,
				result.Coverage,
				result.TotalReturn,
				result.Sharpe,
				result.MaxDrawdown,
				shortPath(csvPath),
			)
		}

		fmt.Println()
		fmt.Println("Notes:")
		fmt.Println("- This baseline exits at next candle close (matches label horizon).")
		fmt.Println("- CSV equity curves are written if --out is set (plotting comes next).")

		return nil
	},
}

func init() {
	paperCommand.Flags().StringVar(&paperSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	paperCommand.Flags().StringVar(&paperTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")
	paperCommand.Flags().StringVar(&paperModelName, "model", "logreg_softmax", "Model name in predictions.model_name")

	paperCommand.Flags().StringVar(&paperThresholds, "thresholds", "0.40,0.45,0.50", "Comma-separated confidence thresholds")
	paperCommand.Flags().Float64Var(&paperFee, "fee", 0.0004, "Fee per side (e.g. 0.0004 = 4 bps)")
	paperCommand.Flags().Float64Var(&paperSlippage, "slippage", 0.0000, "Slippage (round-trip) as decimal")

	paperCommand.Flags().StringVar(&paperOutDir, "out", "reports", "Output directory for equity CSV files (empty disables)")
}

func parseThresholds(value string) ([]float64, error) {
	parts := strings.Split(value, ",")
	var out []float64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var thr float64
		_, err := fmt.Sscanf(p, "%f", &thr)
		if err != nil {
			return nil, fmt.Errorf("invalid threshold %q", p)
		}
		out = append(out, thr)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no thresholds provided")
	}
	return out, nil
}

func shortPath(p string) string {
	if p == "" {
		return ""
	}
	// Keep it readable in terminal
	if len(p) > 28 {
		return "…/" + filepath.Base(p)
	}
	return p
}
