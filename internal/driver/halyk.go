package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// Halyk driver fetches exchange rates from Halyk Bank API.
// Default endpoint: https://back.halykbank.kz/common/currency-history
// We use section "privatePersons" and pick supported currencies (USD, EUR, RUB) vs KZT.
// Response contains history keyed by index (0 - today, 3 - older). We take index 0 only.

type Halyk struct {
	addr       string
	httpClient HTTPClient
}

type halykResponse struct {
	Result bool `json:"result"`
	Data   struct {
		CurrencyHistory map[string]struct {
			Date           string               `json:"date"`
			PrivatePersons map[string]halykPair `json:"privatePersons"`
			LegalPersons   map[string]halykPair `json:"legalPersons"`
			Cards          map[string]halykPair `json:"cards"`
			CrossCourses   map[string]halykPair `json:"crossCourses"`
		} `json:"currencyHistory"`
	} `json:"data"`
}

type halykPair struct {
	Sell float64 `json:"sell"`
	Buy  float64 `json:"buy"`
}

func NewHalyk(addr string, httpClient HTTPClient) *Halyk {
	return &Halyk{addr: addr, httpClient: httpClient}
}

// FetchRates returns latest (index 0) private persons rates for supported currencies.
func (h *Halyk) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
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

	var r halykResponse
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !r.Result {
		return nil, fmt.Errorf("result flag false")
	}
	if len(r.Data.CurrencyHistory) == 0 {
		return nil, fmt.Errorf("empty currency history")
	}

	// take index "0" (latest)
	latest, ok := r.Data.CurrencyHistory["0"]
	if !ok {
		return nil, fmt.Errorf("no latest index 0 in currencyHistory")
	}

	supported := map[string]struct{}{"USD": {}, "EUR": {}, "RUB": {}}
	now := time.Now().UTC()
	var rates []*entity.ExchangeRate
	for pair, v := range latest.PrivatePersons { // keys like USD/KZT
		base := strings.SplitN(pair, "/", 2)[0]
		if _, ok = supported[base]; !ok {
			continue
		}
		rates = append(rates, &entity.ExchangeRate{
			Source:       "Halyk",
			CurrencyCode: base,
			Buy:          strconv.FormatFloat(v.Buy, 'f', -1, 64),
			Sell:         strconv.FormatFloat(v.Sell, 'f', -1, 64),
			CreatedAt:    now,
		})
	}
	if len(rates) == 0 {
		return nil, fmt.Errorf("no supported currency rates found")
	}
	return rates, nil
}
