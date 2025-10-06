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

type rbkMockResponse struct {
	Error int `json:"error"`
	Data  struct {
		Online struct {
			Buy  []rbkMockItem `json:"buy"`
			Sell []rbkMockItem `json:"sell"`
			Date string        `json:"date"`
		} `json:"online"`
	} `json:"data"`
}

type rbkMockItem struct {
	Src    string `json:"src"`
	Dst    string `json:"dst"`
	Scale  string `json:"scale"`
	Amount string `json:"amount"`
}

func TestRBK_FetchRates(t *testing.T) {
	baseResp := rbkMockResponse{}
	baseResp.Data.Online.Buy = []rbkMockItem{
		{Src: "USD", Dst: "KZT", Scale: "1", Amount: "539.50"},
		{Src: "EUR", Dst: "KZT", Scale: "1", Amount: "628.42"},
		{Src: "RUB", Dst: "KZT", Scale: "1", Amount: "6.2620"},
		{Src: "XAU", Dst: "KZT", Scale: "1", Amount: "1781369.00"}, // игнор
	}
	baseResp.Data.Online.Sell = []rbkMockItem{
		{Src: "USD", Dst: "KZT", Scale: "1", Amount: "546.50"},
		{Src: "EUR", Dst: "KZT", Scale: "1", Amount: "637.42"},
		{Src: "RUB", Dst: "KZT", Scale: "1", Amount: "7.3618"},
		{Src: "XAU", Dst: "KZT", Scale: "1", Amount: "1870964.00"}, // игнор
	}
	baseResp.Data.Online.Date = time.Now().Format("2006-01-02")

	t.Run("success", func(t *testing.T) {
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(baseResp)
		})

		d := driver.NewRBK("http://mock", cli)
		rates, err := d.FetchRates(context.Background())
		require.NoError(t, err)
		require.Len(t, rates, 3)
	})

	t.Run("non 200", func(t *testing.T) {
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})

		d := driver.NewRBK("http://mock", cli)
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{"))
		})
		d := driver.NewRBK("http://mock", cli)
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("api error", func(t *testing.T) {
		resp := baseResp
		resp.Error = 1
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		})
		d := driver.NewRBK("http://mock", cli)
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("missing one side", func(t *testing.T) {
		resp := baseResp
		resp.Data.Online.Sell = resp.Data.Online.Sell[:1] // только USD
		resp.Data.Online.Buy = resp.Data.Online.Buy[1:]   // USD buy убрали
		cli := newTestHTTPClient(t, func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		})
		d := driver.NewRBK("http://mock", cli)
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	_ = entity.ExchangeRate{}
}
