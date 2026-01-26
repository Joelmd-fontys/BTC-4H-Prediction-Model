package main

import (
	"BTC-4H-Prediction-Model/internal/exchange"
	"BTC-4H-Prediction-Model/internal/store"
	"context"
	"fmt"
	"io"
	"net/http"
)
import _ "modernc.org/sqlite"

func main() {

	var exchangeName = "binance"
	var symbol = "BTCUSDT"
	var timeframe = "4h"
	var limit = 500
	var dbPath = "btcqd.sqlite"

	// 1) Build URL
	url := exchange.BuildKlinesURL(symbol, timeframe, limit)
	fmt.Println("request:", url)

	// 2) Fetch
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("http get error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("non-200 response:", resp.Status)
		return
	}

	// 3) Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read body error:", err)
		return
	}

	// 4) Parse candles (all 5)
	cs, err := exchange.ParseKlinesToCandles(body, exchangeName, symbol, timeframe)
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}
	fmt.Println("candles parsed:", len(cs))

	// 5) Open DB
	db, err := store.OpenSQLite(dbPath)
	if err != nil {
		fmt.Println("open db error:", err)
		return
	}
	defer db.Close()
	fmt.Println("db opened:", dbPath)

	// 6) Insert ONE candle (first one)
	ctx := context.Background()
	if len(cs) == 0 {
		fmt.Println("no candles to insert")
		return
	}

	if err := store.InsertOneCandle(ctx, db, cs[0]); err != nil {
		fmt.Println("insert error:", err)
		return
	}
	fmt.Println("inserted one candle ts:", cs[0].Timestamp)
}
