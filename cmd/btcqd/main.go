package main

import (
	"BTC-4H-Prediction-Model/internal/exchange"
	"fmt"
	"io"
	"net/http"
)

func main() {
	var exchangeName = "binance"
	var symbol = "BTCUSDT"
	var timeframe = "4h"
	var limit = 5

	url := exchange.BuildKlinesURL(symbol, timeframe, limit)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("encountered an error" + err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println("non-200 response:", resp.Status)
		return
	}
	body, err := io.ReadAll(resp.Body)
	firstCandle, err := exchange.ParseFirstKlineToCandle(body, exchangeName, symbol, timeframe)
	fmt.Println(firstCandle)
}
