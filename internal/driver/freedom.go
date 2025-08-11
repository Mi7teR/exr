package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// Freedom driver fetches exchange rates from Freedom Bank API.
// Endpoint: https://bankffin.kz/api/exchange-rates/getRates
// We take section "cash" and pick supported currencies (USD, EUR, RUB) where sellCode == KZT.

type Freedom struct {
	addr       string
	httpClient HTTPClient
}

type freedomResponse struct {
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
	Data    struct {
		Cash    []freedomItem `json:"cash"`
		Mobile  []freedomItem `json:"mobile"`
		NonCash []freedomItem `json:"non_cash"`
	} `json:"data"`
	Status int `json:"status"`
}

type freedomItem struct {
	BuyCode  string `json:"buyCode"`
	SellCode string `json:"sellCode"`
	BuyRate  string `json:"buyRate"`
	SellRate string `json:"sellRate"`
}

// NewFreedom creates a new Freedom driver.
func NewFreedom(addr string, httpClient HTTPClient) *Freedom {
	return &Freedom{addr: addr, httpClient: httpClient}
}

// FetchRates fetches cash exchange rates for supported currencies.
func (f *Freedom) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.addr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var fr freedomResponse
	if err = json.NewDecoder(resp.Body).Decode(&fr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !fr.Success {
		return nil, fmt.Errorf("freedom api success false")
	}
	if fr.Status != 200 && fr.Status != 0 { // some responses might omit status
		return nil, fmt.Errorf("freedom api status %d", fr.Status)
	}

	supported := map[string]struct{}{"USD": {}, "EUR": {}, "RUB": {}}
	now := time.Now().UTC()
	var rates []*entity.ExchangeRate
	for _, it := range fr.Data.Cash {
		if it.SellCode != "KZT" {
			continue
		}
		if _, ok := supported[it.BuyCode]; !ok {
			continue
		}
		rates = append(rates, &entity.ExchangeRate{
			Source:       "Freedom",
			CurrencyCode: it.BuyCode,
			Buy:          it.BuyRate,
			Sell:         it.SellRate,
			CreatedAt:    now,
		})
	}
	if len(rates) == 0 {
		return nil, fmt.Errorf("no supported currency rates found")
	}
	return rates, nil
}
