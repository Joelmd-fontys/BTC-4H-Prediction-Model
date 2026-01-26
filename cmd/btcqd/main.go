package main

import (
	"BTC-4H-Prediction-Model/internal/exchange"
	"fmt"
)

func main() {
	fmt.Printf(exchange.BuildKlinesURL("BTCUSDT", "4h", 5))
}
