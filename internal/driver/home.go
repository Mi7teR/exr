package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// Home driver (home.kz) fetches exchange rates.
// Primary endpoint: https://home.kz/strapi/api/getCurrency?locale=ru
// Response contains currency nodes (usd/eur/rub) with buy/sell text fields.
// Driver keeps backward compatibility with legacy /api/public/getCurrency payload.

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

type homeStrapiCurrency struct {
	CurrencyID string `json:"currency_id"`
	RateBuy    string `json:"rate_buy"`
	RateSell   string `json:"rate_sell"`
	Updated    string `json:"upd_time"`
}

type homeStrapiResponse struct {
	Date string              `json:"date"`
	USD  *homeStrapiCurrency `json:"usd"`
	Euro *homeStrapiCurrency `json:"euro"`
	Rub  *homeStrapiCurrency `json:"rub"`
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if rates := parseHomeStrapi(body); len(rates) > 0 {
		return rates, nil
	}

	var legacy homeResponse
	if err = json.Unmarshal(body, &legacy); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	now := time.Now().UTC()
	var out []*entity.ExchangeRate
	supported := map[string]string{"1": "USD", "17": "EUR", "16": "RUB"}
	for _, c := range legacy.Currency {
		code, ok := supported[c.ID]
		if !ok {
			continue
		}
		createdAt := now
		if t, parseErr := parseHomeTime(c.Updated); parseErr == nil {
			createdAt = t
		}
		out = append(out, &entity.ExchangeRate{
			Source:       "HomeKZ",
			CurrencyCode: code,
			Buy:          c.Buy,
			Sell:         c.Sell,
			CreatedAt:    createdAt,
		})
	}
	if len(out) == 0 {
		return nil, errors.New("no supported currency rates found")
	}
	return out, nil
}

func parseHomeStrapi(body []byte) []*entity.ExchangeRate {
	var sr homeStrapiResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil
	}

	mapping := map[string]*homeStrapiCurrency{
		"USD": sr.USD,
		"EUR": sr.Euro,
		"RUB": sr.Rub,
	}
	var rates []*entity.ExchangeRate
	for code, item := range mapping {
		if item == nil {
			continue
		}
		if item.RateBuy == "" && item.RateSell == "" {
			continue
		}
		createdAt := time.Now().UTC()
		if t, err := parseHomeTime(item.Updated); err == nil {
			createdAt = t
		}
		rates = append(rates, &entity.ExchangeRate{
			Source:       "HomeKZ",
			CurrencyCode: code,
			Buy:          item.RateBuy,
			Sell:         item.RateSell,
			CreatedAt:    createdAt,
		})
	}
	return rates
}

func parseHomeTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty time")
	}
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05.000-07:00",
		"2006-01-02T15:04:05-07:00",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time: %s", value)
}
