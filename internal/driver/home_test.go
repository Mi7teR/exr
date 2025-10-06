package driver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Mi7teR/exr/internal/driver"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/stretchr/testify/require"
)

type homeStrapiMock struct {
	Date string `json:"date"`
	USD  struct {
		CurrencyID string `json:"currency_id"`
		RateBuy    string `json:"rate_buy"`
		RateSell   string `json:"rate_sell"`
		Updated    string `json:"upd_time"`
	} `json:"usd"`
	Euro struct {
		CurrencyID string `json:"currency_id"`
		RateBuy    string `json:"rate_buy"`
		RateSell   string `json:"rate_sell"`
		Updated    string `json:"upd_time"`
	} `json:"euro"`
	Rub struct {
		CurrencyID string `json:"currency_id"`
		RateBuy    string `json:"rate_buy"`
		RateSell   string `json:"rate_sell"`
		Updated    string `json:"upd_time"`
	} `json:"rub"`
}

type homeMockResponse struct {
	Currency []struct {
		ID      string `json:"p_curr_id"`
		Buy     string `json:"p_rate_buy"`
		Sell    string `json:"p_rate_sell"`
		Updated string `json:"p_last_upd"`
	} `json:"currency"`
}

func TestHome_FetchRates_StrapiPayload(t *testing.T) {
	mock := homeStrapiMock{Date: "06.10.2025"}
	mock.USD = struct {
		CurrencyID string `json:"currency_id"`
		RateBuy    string `json:"rate_buy"`
		RateSell   string `json:"rate_sell"`
		Updated    string `json:"upd_time"`
	}{"1", "543.5", "545.5", "2025-10-06T14:36:20.000+05:00"}
	mock.Euro = struct {
		CurrencyID string `json:"currency_id"`
		RateBuy    string `json:"rate_buy"`
		RateSell   string `json:"rate_sell"`
		Updated    string `json:"upd_time"`
	}{"17", "633.43", "637.43", "2025-10-06T14:36:20.000+05:00"}

	cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(mock)
	})

	d := driver.NewHome("http://mock", cli)
	rates, err := d.FetchRates(context.Background())
	require.NoError(t, err)
	require.Len(t, rates, 2)

	currencies := map[string]*entity.ExchangeRate{}
	for _, r := range rates {
		currencies[r.CurrencyCode] = r
	}
	require.Contains(t, currencies, "USD")
	require.Equal(t, "543.5", currencies["USD"].Buy)
	require.Greater(t, currencies["USD"].CreatedAt, time.Time{})
}

func TestHome_FetchRates_LegacyPayload(t *testing.T) {
	legacy := struct {
		Currency []struct {
			ID      string `json:"p_curr_id"`
			Buy     string `json:"p_rate_buy"`
			Sell    string `json:"p_rate_sell"`
			Updated string `json:"p_last_upd"`
		} `json:"currency"`
	}{
		Currency: []struct {
			ID      string `json:"p_curr_id"`
			Buy     string `json:"p_rate_buy"`
			Sell    string `json:"p_rate_sell"`
			Updated string `json:"p_last_upd"`
		}{
			{ID: "1", Buy: "532.8", Sell: "534.8", Updated: "2024-05-01T12:00:00+05:00"},
			{ID: "20", Buy: "10", Sell: "11", Updated: "2024-05-01T12:00:00+05:00"},
		},
	}

	cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(legacy)
	})

	d := driver.NewHome("http://mock", cli)
	rates, err := d.FetchRates(context.Background())
	require.NoError(t, err)
	require.Len(t, rates, 1)
	require.Equal(t, "USD", rates[0].CurrencyCode)
	require.Equal(t, "532.8", rates[0].Buy)
}

func TestHome_FetchRates_Errors(t *testing.T) {
	t.Run("non-200", func(t *testing.T) {
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})
		d := driver.NewHome("http://mock", cli)
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		d := driver.NewHome("http://mock", cli)
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
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		})
		d := driver.NewHome("http://mock", cli)
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	_ = entity.ExchangeRate{}
	_ = time.Now()
}
