package exchange

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type BinanceClient struct {
	HTTPClient *http.Client
}

func NewBinanceClient() BinanceClient {
	return BinanceClient{HTTPClient: http.DefaultClient}
}

func (client BinanceClient) Get(context context.Context, requestURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(context, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance non-200: %s body=%s", response.Status, string(body))
	}

	return body, nil
}
