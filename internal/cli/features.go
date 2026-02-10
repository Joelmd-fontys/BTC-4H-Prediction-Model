package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/features"
	"btc-4h-prediction-model/internal/store"
)

var featuresSymbol string
var featuresTimeframe string

var featuresCommand = &cobra.Command{
	Use:   "features",
	Short: "Compute features from candles and store them in the database",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		db, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer db.Close()

		candleSeries, err := store.LoadCandlesOrdered(ctx, db, "binance", featuresSymbol, featuresTimeframe)
		if err != nil {
			return err
		}
		if len(candleSeries) == 0 {
			return fmt.Errorf("no candles found for %s %s", featuresSymbol, featuresTimeframe)
		}

		featureRows, err := features.BuildFeaturesFromCandles(candleSeries)
		if err != nil {
			return err
		}

		if err := store.UpsertFeatures(ctx, db, featureRows); err != nil {
			return err
		}

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", featuresSymbol)
		fmt.Println("timeframe:", featuresTimeframe)
		fmt.Println("features rows upserted:", len(featureRows))
		return nil
	},
}

func init() {
	featuresCommand.Flags().StringVar(&featuresSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	featuresCommand.Flags().StringVar(&featuresTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")
}
