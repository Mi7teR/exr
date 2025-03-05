package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Mi7teR/exr/internal/entity"
)

const (
	kaspiRequestBody     = `{"use_type":"32","currency_codes":["USD","EUR"],"rate_types":["SALE","BUY"]}`
	kaspiHeaderGLanguage = "gLanguage"
	kaspiHeaderGSystem   = "gSystem"
)

type Kaspi struct {
	addr       string
	httpClient HTTPClient
}

type KaspiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Body    []struct {
		Currency string `json:"currency"`
		Buy      int    `json:"buy"`
		Sale     int    `json:"sale"`
	} `json:"body"`
}

// NewKaspi creates a new Kaspi driver.
// Default addr is "https://guide.kaspi.kz/client/api/intgr/currency/rate/aggregate".
// Seems that addr works only from KZ location.
func NewKaspi(addr string, httpClient HTTPClient) *Kaspi {
	return &Kaspi{
		addr:       addr,
		httpClient: httpClient,
	}
}

// FetchRates fetches exchange rates from the Kaspi API.
func (k *Kaspi) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
	reqBody := bytes.NewBufferString(kaspiRequestBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.addr, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(kaspiHeaderGLanguage, "ru") //nolint: canonicalheader // should use no canonical header
	req.Header.Set(kaspiHeaderGSystem, "kkz")  //nolint: canonicalheader // should use no canonical header

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var res KaspiResponse

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	if res.Status != "OK" || res.Message != "OK" {
		return nil, fmt.Errorf("Kaspi API returned error: %s", res.Message)
	}

	var rates []*entity.ExchangeRate

	for _, item := range res.Body {
		rates = append(rates, &entity.ExchangeRate{
			Source:       "Kaspi",
			CurrencyCode: item.Currency,
			Buy:          strconv.FormatInt(int64(item.Buy), 10),
			Sell:         strconv.FormatInt(int64(item.Sale), 10),
		})
	}

	return rates, nil
}
