package exchange

import (
	"net/url"
	"strconv"
)

func BuildKlinesURL(symbol string, interval string, limit int, startTimeMillis int64, endTimeMillis int64) string {
	baseURL := "https://api.binance.com"
	path := "/api/v3/klines"

	query := url.Values{}
	query.Set("symbol", symbol)
	query.Set("interval", interval)
	query.Set("limit", strconv.Itoa(limit))

	if startTimeMillis > 0 {
		query.Set("startTime", strconv.FormatInt(startTimeMillis, 10))
	}
	if endTimeMillis > 0 {
		query.Set("endTime", strconv.FormatInt(endTimeMillis, 10))
	}

	return baseURL + path + "?" + query.Encode()
}
