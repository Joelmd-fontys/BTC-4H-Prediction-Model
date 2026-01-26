package exchange

import (
	"net/url"
	"strconv"
)

func BuildKlinesURL(symbol string, interval string, limit int) string {
	url := url.URL{
		Scheme: "https",
		Host:   "api.binance.com",
		Path:   "/api/v3/klines",
	}

	query := url.Query()
	query.Set("symbol", symbol)
	query.Set("interval", interval)
	query.Set("limit", strconv.Itoa(limit))

	url.RawQuery = query.Encode()
	return url.String()
}
