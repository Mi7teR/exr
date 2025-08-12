package driver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mi7teR/exr/internal/driver"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/stretchr/testify/require"
)

type pricePair struct {
	Sell float64 `json:"sell"`
	Buy  float64 `json:"buy"`
}

type halykMockResp struct {
	Result bool `json:"result"`
	Data   struct {
		CurrencyHistory map[string]struct {
			Date           string               `json:"date"`
			PrivatePersons map[string]pricePair `json:"privatePersons"`
		} `json:"currencyHistory"`
	} `json:"data"`
}

func TestHalyk_FetchRates(t *testing.T) {
	baseResp := halykMockResp{Result: true}
	baseResp.Data.CurrencyHistory = map[string]struct {
		Date           string               `json:"date"`
		PrivatePersons map[string]pricePair `json:"privatePersons"`
	}{
		"0": {
			Date: time.Now().Format("2006-01-02"),
			PrivatePersons: map[string]pricePair{
				"USD/KZT": {Sell: 544.6, Buy: 537.6},
				"EUR/KZT": {Sell: 635.26, Buy: 625.76},
				"RUB/KZT": {Sell: 7.07, Buy: 6.57},
				"XAU/USD": {Sell: 3568.7, Buy: 3178.16}, // игнор
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(baseResp)
		}))
		defer server.Close()

		d := driver.NewHalyk(server.URL, server.Client())
		rates, err := d.FetchRates(context.Background())
		require.NoError(t, err)
		require.Len(t, rates, 3)
	})

	t.Run("non 200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()

		d := driver.NewHalyk(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()
		d := driver.NewHalyk(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("empty history", func(t *testing.T) {
		emptyResp := halykMockResp{Result: true}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(emptyResp)
		}))
		defer server.Close()
		d := driver.NewHalyk(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("no supported", func(t *testing.T) {
		resp := halykMockResp{Result: true}
		resp.Data.CurrencyHistory = map[string]struct {
			Date           string               `json:"date"`
			PrivatePersons map[string]pricePair `json:"privatePersons"`
		}{
			"0": {
				Date:           time.Now().Format("2006-01-02"),
				PrivatePersons: map[string]pricePair{"XAU/USD": {Sell: 1, Buy: 2}},
			},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		d := driver.NewHalyk(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	_ = entity.ExchangeRate{} // silence unused import if CI caching
}
