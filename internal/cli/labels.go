package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/labels"
	"btc-4h-prediction-model/internal/store"
)

var labelsSymbol string
var labelsTimeframe string
var labelsThresholdB float64

var labelsCommand = &cobra.Command{
	Use:   "labels",
	Short: "Compute forward-return labels (UP/DOWN/NO_TRADE) and store them in the database",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		db, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer db.Close()

		candleSeries, err := store.LoadCandlesOrdered(ctx, db, "binance", labelsSymbol, labelsTimeframe)
		if err != nil {
			return err
		}
		if len(candleSeries) < 2 {
			return fmt.Errorf("not enough candles to label (%d)", len(candleSeries))
		}

		labelRows, err := labels.BuildForwardReturnLabels(candleSeries, labelsThresholdB)
		if err != nil {
			return err
		}

		if err := store.UpsertLabels(ctx, db, labelRows); err != nil {
			return err
		}

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", labelsSymbol)
		fmt.Println("timeframe:", labelsTimeframe)
		fmt.Println("b:", labelsThresholdB)
		fmt.Println("labels upserted:", len(labelRows))
		fmt.Println("note: last candle has no label (no future candle)")

		return nil
	},
}

func init() {
	labelsCommand.Flags().StringVar(&labelsSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	labelsCommand.Flags().StringVar(&labelsTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")
	labelsCommand.Flags().Float64Var(&labelsThresholdB, "b", 0.006, "Threshold b as decimal return (e.g. 0.006 = 0.6%)")
}
