package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
// Response may contain history either as a map indexed by strings ("0", "1", ...) or as an array.

type Halyk struct {
	addr       string
	httpClient HTTPClient
}

type halykHistoryEntry struct {
	Date           string               `json:"date"`
	PrivatePersons map[string]halykPair `json:"privatePersons"`
	LegalPersons   map[string]halykPair `json:"legalPersons"`
	Cards          map[string]halykPair `json:"cards"`
	CrossCourses   map[string]halykPair `json:"crossCourses"`
}

type currencyHistory struct {
	byIndex map[string]halykHistoryEntry
	list    []halykHistoryEntry
}

func (c *currencyHistory) UnmarshalJSON(b []byte) error {
	bb := bytes.TrimSpace(b)
	if len(bb) == 0 {
		return nil
	}
	switch bb[0] {
	case '{':
		var m map[string]halykHistoryEntry
		if err := json.Unmarshal(bb, &m); err != nil {
			return err
		}
		c.byIndex = m
	case '[':
		var l []halykHistoryEntry
		if err := json.Unmarshal(bb, &l); err != nil {
			return err
		}
		c.list = l
	default:
		return fmt.Errorf("unexpected currencyHistory JSON")
	}
	return nil
}

func (c currencyHistory) Latest() (halykHistoryEntry, bool) {
	if c.byIndex != nil {
		if e, ok := c.byIndex["0"]; ok {
			return e, true
		}
	}
	if len(c.list) > 0 {
		return c.list[0], true
	}
	return halykHistoryEntry{}, false
}

type halykResponse struct {
	Result bool `json:"result"`
	Data   struct {
		CurrencyHistory currencyHistory `json:"currencyHistory"`
	} `json:"data"`
}

type halykPair struct {
	Sell float64 `json:"sell"`
	Buy  float64 `json:"buy"`
}

func NewHalyk(addr string, httpClient HTTPClient) *Halyk {
	return &Halyk{addr: addr, httpClient: httpClient}
}

const (
	pairSeparator = "/"
	splitLimit    = 2
)

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
		return nil, errors.New("result flag false")
	}

	latest, ok := r.Data.CurrencyHistory.Latest()
	if !ok {
		return nil, errors.New("empty currency history")
	}

	supported := map[string]struct{}{"USD": {}, "EUR": {}, "RUB": {}}
	now := time.Now().UTC()
	var rates []*entity.ExchangeRate
	for pair, v := range latest.PrivatePersons { // keys like USD/KZT
		base := strings.SplitN(pair, pairSeparator, splitLimit)[0]
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
		return nil, errors.New("no supported currency rates found")
	}
	return rates, nil
}
