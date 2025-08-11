package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// RBK driver fetches exchange rates from RBK Bank API.
// Default endpoint: https://backend.bankrbk.kz/api/v1/modules/exchange_rates/data
// We use section "online" and pick supported currencies (USD, EUR, RUB) where dst == KZT.

type RBK struct {
	addr       string
	httpClient HTTPClient
}

type rbkResponse struct {
	Error int `json:"error"`
	Data  struct {
		Online struct {
			Buy  []rbkItem `json:"buy"`
			Sell []rbkItem `json:"sell"`
			Date string    `json:"date"`
		} `json:"online"`
		Branch struct { // kept for potential future use
			Buy  []rbkItem `json:"buy"`
			Sell []rbkItem `json:"sell"`
			Date string    `json:"date"`
		} `json:"branch"`
	} `json:"data"`
}

type rbkItem struct {
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Scale  string `json:"scale"`
	Amount string `json:"amount"`
}

// NewRBK creates RBK driver.
func NewRBK(addr string, httpClient HTTPClient) *RBK {
	return &RBK{addr: addr, httpClient: httpClient}
}

// FetchRates returns latest online rates for supported currencies.
func (r *RBK) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.addr, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var rr rbkResponse
	if err = json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if rr.Error != 0 {
		return nil, fmt.Errorf("rbk api error code: %d", rr.Error)
	}

	supported := map[string]struct{}{"USD": {}, "EUR": {}, "RUB": {}}
	buyMap := make(map[string]string)
	sellMap := make(map[string]string)

	for _, it := range rr.Data.Online.Buy {
		if it.Dst != "KZT" { // skip cross / metals
			continue
		}
		if _, ok := supported[it.Src]; !ok {
			continue
		}
		buyMap[it.Src] = it.Amount
	}
	for _, it := range rr.Data.Online.Sell {
		if it.Dst != "KZT" {
			continue
		}
		if _, ok := supported[it.Src]; !ok {
			continue
		}
		sellMap[it.Src] = it.Amount
	}

	now := time.Now().UTC()
	var rates []*entity.ExchangeRate
	for cur := range supported {
		buy, bok := buyMap[cur]
		sell, sok := sellMap[cur]
		if !bok || !sok {
			continue // skip if one side missing
		}
		rates = append(rates, &entity.ExchangeRate{
			Source:       "RBK",
			CurrencyCode: cur,
			Buy:          buy,
			Sell:         sell,
			CreatedAt:    now,
		})
	}
	if len(rates) == 0 {
		return nil, fmt.Errorf("no supported currency rates found")
	}
	return rates, nil
}
