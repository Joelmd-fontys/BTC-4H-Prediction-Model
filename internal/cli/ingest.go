package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"btc-4h-prediction-model/internal/exchange"
	"btc-4h-prediction-model/internal/store"
)

var ingestSymbol string
var ingestTimeframe string
var ingestDays int

var ingestCommand = &cobra.Command{
	Use:   "ingest",
	Short: "Fetch Binance candles and upsert into the database",
	RunE: func(command *cobra.Command, args []string) error {
		ctx := context.Background()

		database, err := store.OpenSQLite(databasePath)
		if err != nil {
			return err
		}
		defer database.Close()

		if ingestDays <= 0 {
			return fmt.Errorf("--days must be > 0")
		}

		// v0.1: look back N days from now
		now := time.Now().UTC()
		startTimeMillis := now.Add(time.Duration(-ingestDays) * 24 * time.Hour).UnixMilli()

		binanceClient := exchange.BinanceClient{HTTPClient: http.DefaultClient}

		candlesFetched, err := binanceClient.FetchKlinesPaginated(
			ctx,
			ingestSymbol,
			ingestTimeframe,
			startTimeMillis,
			0, // no endTime
		)
		if err != nil {
			return err
		}

		if err := store.UpsertCandles(ctx, database, candlesFetched); err != nil {
			return err
		}

		fmt.Println("db:", databasePath)
		fmt.Println("exchange: binance")
		fmt.Println("symbol:", ingestSymbol)
		fmt.Println("timeframe:", ingestTimeframe)
		fmt.Println("days:", ingestDays)
		fmt.Println("fetched:", len(candlesFetched))
		fmt.Println("upserted:", len(candlesFetched))

		return nil
	},
}

func init() {
	ingestCommand.Flags().StringVar(&ingestSymbol, "symbol", "BTCUSDT", "Symbol (e.g. BTCUSDT)")
	ingestCommand.Flags().StringVar(&ingestTimeframe, "timeframe", "4h", "Timeframe (e.g. 4h)")
	ingestCommand.Flags().IntVar(&ingestDays, "days", 30, "Lookback window in days")
}
