package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// Home driver (home.kz) fetches exchange rates.
// Endpoint: https://home.kz/api/public/getCurrency
// We map p_curr_id to currency codes of interest (USD, EUR, RUB) vs KZT.

type Home struct {
	addr       string
	httpClient HTTPClient
}

type homeResponse struct {
	Currency []struct {
		ID      string `json:"p_curr_id"`
		Buy     string `json:"p_rate_buy"`
		Sell    string `json:"p_rate_sell"`
		Updated string `json:"p_last_upd"`
	} `json:"currency"`
}

func homeSupportedMap() map[string]string { // id -> currency code
	return map[string]string{
		"1":  "USD",
		"17": "EUR",
		"16": "RUB",
	}
}

// NewHome creates driver for home.kz.
func NewHome(addr string, httpClient HTTPClient) *Home {
	return &Home{addr: addr, httpClient: httpClient}
}

// FetchRates returns supported currency rates.
func (h *Home) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.addr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var r homeResponse
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	now := time.Now().UTC()
	var rates []*entity.ExchangeRate
	for _, c := range r.Currency {
		code, ok := homeSupportedMap()[c.ID]
		if !ok {
			continue
		}
		rates = append(rates, &entity.ExchangeRate{
			Source:       "HomeKZ",
			CurrencyCode: code,
			Buy:          c.Buy,
			Sell:         c.Sell,
			CreatedAt:    now,
		})
	}
	if len(rates) == 0 {
		return nil, errors.New("no supported currency rates found")
	}
	return rates, nil
}
