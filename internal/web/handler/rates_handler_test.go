package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Mi7teR/exr/internal/application/logger"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/service/exrate"
)

type mockLogger struct{}

func (m *mockLogger) Debug(string, ...any)      {}
func (m *mockLogger) Info(string, ...any)       {}
func (m *mockLogger) Warn(string, ...any)       {}
func (m *mockLogger) Error(string, ...any)      {}
func (m *mockLogger) With(...any) logger.Logger { return m }

type mockRatesService struct {
	getRatesFunc func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
}

func (m *mockRatesService) GetRates(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
	if m.getRatesFunc != nil {
		return m.getRatesFunc(ctx, filter)
	}
	return nil, nil
}

func TestRatesHandler_FullPage(t *testing.T) {
	fixed := time.Date(2024, time.January, 1, 10, 0, 0, 0, time.UTC)
	svc := &mockRatesService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return []*entity.ExchangeRate{
				{
					CurrencyCode:   "USD",
					Buy:            "490.50",
					Sell:           "495.00",
					Source:         "Kaspi",
					CreatedAt:      fixed,
					BuyChangePrev:  1.5,
					SellChangePrev: 2.0,
				},
			}, nil
		},
	}

	handler := NewRatesHandler(&mockLogger{}, svc)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/c/usd", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("unexpected content-type %s", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected body content")
	}

	body := rec.Body.String()
	expectedISO := fixed.UTC().Format(time.RFC3339)
	if !strings.Contains(body, "data-initial-timestamp=\""+expectedISO+"\"") {
		t.Fatalf("expected initial timestamp %s in body", expectedISO)
	}
}

func TestRatesHandler_IndexPageUsesDefaultCurrency(t *testing.T) {
	fixed := time.Date(2024, time.March, 5, 9, 0, 0, 0, time.UTC)
	svc := &mockRatesService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return []*entity.ExchangeRate{{
				CurrencyCode: "USD",
				Buy:          "500",
				Sell:         "505",
				Source:       "Test",
				CreatedAt:    fixed,
			}}, nil
		},
	}

	handler := NewRatesHandler(&mockLogger{}, svc)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	expectedISO := fixed.UTC().Format(time.RFC3339)
	if !strings.Contains(rec.Body.String(), "data-initial-timestamp=\""+expectedISO+"\"") {
		t.Fatalf("expected initial timestamp %s", expectedISO)
	}
}

func TestRatesHandler_HTMX(t *testing.T) {
	fixed := time.Date(2024, time.January, 2, 12, 30, 0, 0, time.UTC)
	svc := &mockRatesService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return []*entity.ExchangeRate{
				{
					CurrencyCode:   "USD",
					Buy:            "490.50",
					Sell:           "495.00",
					Source:         "Kaspi",
					CreatedAt:      fixed,
					BuyChangePrev:  1.5,
					SellChangePrev: 2.0,
				},
			}, nil
		},
	}

	handler := NewRatesHandler(&mockLogger{}, svc)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/c/usd", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Last-Updated") == "" {
		t.Fatal("expected X-Last-Updated header")
	}
	body := rec.Body.String()
	expectedISO := fixed.UTC().Format(time.RFC3339)
	if !strings.Contains(body, "data-last-updated=\""+expectedISO+"\"") {
		t.Fatalf("expected data-last-updated attribute %s", expectedISO)
	}
}

func TestRatesHandler_InvalidCurrency(t *testing.T) {
	handler := NewRatesHandler(&mockLogger{}, &mockRatesService{})
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/c/gbp", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
