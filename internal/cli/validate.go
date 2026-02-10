package cli

import (
	"btc-4h-prediction-model/internal/candles"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/ingest"
	"btc-4h-prediction-model/internal/store"
)

var validateSymbol string
var validateTimeframe string

var validateCommand = &cobra.Command{
	Use:   "validate",
	Short: "Validate candle continuity (gaps / ordering) in the database",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		database, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer database.Close()

		intervalMillis, err := candles.TimeframeToMillis(validateTimeframe)
		if err != nil {
			return err
		}

		validationResult, err := ingest.ValidateCandleContinuity(
			ctx,
			database,
			"binance",
			validateSymbol,
			validateTimeframe,
			intervalMillis,
		)
		if err != nil {
			return err
		}

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", validateSymbol)
		fmt.Println("timeframe:", validateTimeframe)
		fmt.Println("count:", validationResult.Count)
		fmt.Println("first ts:", validationResult.FirstTS)
		fmt.Println("last ts:", validationResult.LastTS)
		fmt.Println("gaps:", len(validationResult.Gaps))

		maxGapsToPrint := 5
		for i, gap := range validationResult.Gaps {
			if i >= maxGapsToPrint {
				break
			}
			fmt.Printf(
				"gap %d: prev=%d expected=%d actual=%d missingIntervals=%d\n",
				i+1, gap.PreviousTS, gap.ExpectedTS, gap.ActualTS, gap.Missing,
			)
		}

		return nil
	},
}

func init() {
	validateCommand.Flags().StringVar(&validateSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	validateCommand.Flags().StringVar(&validateTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")
}
