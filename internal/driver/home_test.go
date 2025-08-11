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

type homeMockResponse struct {
	Currency []struct {
		ID      string `json:"p_curr_id"`
		Buy     string `json:"p_rate_buy"`
		Sell    string `json:"p_rate_sell"`
		Updated string `json:"p_last_upd"`
	} `json:"currency"`
}

func TestHome_FetchRates(t *testing.T) {
	base := homeMockResponse{Currency: []struct {
		ID      string `json:"p_curr_id"`
		Buy     string `json:"p_rate_buy"`
		Sell    string `json:"p_rate_sell"`
		Updated string `json:"p_last_upd"`
	}{
		{ID: "1", Buy: "532.8", Sell: "534.8"},
		{ID: "17", Buy: "624.11", Sell: "627.11"},
		{ID: "16", Buy: "6.35", Sell: "6.65"},
		{ID: "20", Buy: "66.72", Sell: "78.72"}, // игнор
	}}

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(base)
		}))
		defer server.Close()
		d := driver.NewHome(server.URL, server.Client())
		rates, err := d.FetchRates(context.Background())
		require.NoError(t, err)
		require.Len(t, rates, 3)
	})

	t.Run("non 200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()
		d := driver.NewHome(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()
		d := driver.NewHome(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("no supported", func(t *testing.T) {
		resp := homeMockResponse{Currency: []struct {
			ID      string `json:"p_curr_id"`
			Buy     string `json:"p_rate_buy"`
			Sell    string `json:"p_rate_sell"`
			Updated string `json:"p_last_upd"`
		}{{ID: "20", Buy: "1", Sell: "2"}}}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		d := driver.NewHome(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	_ = entity.ExchangeRate{}
	_ = time.Now()
}
