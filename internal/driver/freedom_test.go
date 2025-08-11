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

type freedomMockResponse struct {
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
	Data    struct {
		Cash []struct {
			BuyCode  string `json:"buyCode"`
			SellCode string `json:"sellCode"`
			BuyRate  string `json:"buyRate"`
			SellRate string `json:"sellRate"`
		} `json:"cash"`
	} `json:"data"`
	Status int `json:"status"`
}

func TestFreedom_FetchRates(t *testing.T) {
	base := freedomMockResponse{Success: true, Status: 200}
	base.Data.Cash = []struct {
		BuyCode  string `json:"buyCode"`
		SellCode string `json:"sellCode"`
		BuyRate  string `json:"buyRate"`
		SellRate string `json:"sellRate"`
	}{
		{BuyCode: "USD", SellCode: "KZT", BuyRate: "539.00", SellRate: "546.00"},
		{BuyCode: "RUB", SellCode: "KZT", BuyRate: "6.55", SellRate: "7.05"},
		{BuyCode: "EUR", SellCode: "KZT", BuyRate: "629.50", SellRate: "636.50"},
		{BuyCode: "USD", SellCode: "RUB", BuyRate: "77.85", SellRate: "82.64"}, // игнор
	}

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(base)
		}))
		defer server.Close()
		d := driver.NewFreedom(server.URL, server.Client())
		rates, err := d.FetchRates(context.Background())
		require.NoError(t, err)
		require.Len(t, rates, 3)
	})

	t.Run("non 200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer server.Close()
		d := driver.NewFreedom(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("{"))
		}))
		defer server.Close()
		d := driver.NewFreedom(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("api success false", func(t *testing.T) {
		resp := base
		resp.Success = false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		d := driver.NewFreedom(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	t.Run("no supported", func(t *testing.T) {
		resp := freedomMockResponse{Success: true, Status: 200}
		resp.Data.Cash = []struct {
			BuyCode  string `json:"buyCode"`
			SellCode string `json:"sellCode"`
			BuyRate  string `json:"buyRate"`
			SellRate string `json:"sellRate"`
		}{{BuyCode: "AED", SellCode: "KZT", BuyRate: "1", SellRate: "2"}}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()
		d := driver.NewFreedom(server.URL, server.Client())
		_, err := d.FetchRates(context.Background())
		require.Error(t, err)
	})

	_ = entity.ExchangeRate{}
	_ = time.Now()
}
