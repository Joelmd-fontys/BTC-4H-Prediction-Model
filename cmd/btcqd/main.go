package main

import (
	"BTC-4H-Prediction-Model/internal/exchange"
	"BTC-4H-Prediction-Model/internal/ingest"
	"BTC-4H-Prediction-Model/internal/store"
	"context"
	"database/sql"
	"fmt"
	"time"
)

func main() {
	context := context.Background()

	// Use ONE db path consistently (recommend storing DB in /data)
	databasePath := "btcqd.sqlite"

	database, err := store.OpenSQLite(databasePath)
	if err != nil {
		fmt.Println("open db error:", err)
		return
	}
	defer func(database *sql.DB) {
		err := database.Close()
		if err != nil {
		}
	}(database)
	fmt.Println("db opened:", databasePath)

	// Fetch: last 30 days of 4H candles
	now := time.Now().UTC()
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
	startTimeMillis := thirtyDaysAgo.UnixMilli()
	endTimeMillis := int64(0) // 0 means "no endTime filter"

	binanceClient := exchange.NewBinanceClient()

	candlesFetched, err := binanceClient.FetchKlinesPaginated(
		context,
		"BTCUSDT",
		"4h",
		startTimeMillis,
		endTimeMillis,
	)
	if err != nil {
		fmt.Println("fetch error:", err)
		return
	}

	if len(candlesFetched) == 0 {
		fmt.Println("no candles fetched")
		return
	}

	fmt.Println("candles fetched:", len(candlesFetched))
	fmt.Printf("first ts: %d\n", candlesFetched[0].Timestamp)
	fmt.Printf("last  ts: %d\n", candlesFetched[len(candlesFetched)-1].Timestamp)

	// Store: upsert all
	if err := store.UpsertCandles(context, database, candlesFetched); err != nil {
		fmt.Println("upsert error:", err)
		return
	}
	fmt.Println("upserted:", len(candlesFetched))

	// Verify count
	count, err := countCandles(context, database)
	if err != nil {
		fmt.Println("count error:", err)
		return
	}
	fmt.Println("candles in db:", count)

	validationResult, err := ingest.ValidateCandleContinuity(
		context,
		database,
		"binance",
		"BTCUSDT",
		"4h",
		4*60*60*1000, // 4H in ms
	)
	if err != nil {
		fmt.Println("validation error:", err)
		return
	}

	fmt.Println("validation:")
	fmt.Println("  count:", validationResult.Count)
	fmt.Println("  first ts:", validationResult.FirstTS)
	fmt.Println("  last ts:", validationResult.LastTS)
	fmt.Println("  gaps:", len(validationResult.Gaps))

	// Print first few gaps (don’t spam)
	maxGapsToPrint := 5
	for i, gap := range validationResult.Gaps {
		if i >= maxGapsToPrint {
			break
		}
		fmt.Printf("  gap %d: prev=%d expected=%d actual=%d missingIntervals=%d\n",
			i+1, gap.PreviousTS, gap.ExpectedTS, gap.ActualTS, gap.Missing)
	}
}

func countCandles(context context.Context, database *sql.DB) (int64, error) {
	var count int64
	err := database.QueryRowContext(context, `SELECT COUNT(*) FROM candles;`).Scan(&count)
	return count, err
}
