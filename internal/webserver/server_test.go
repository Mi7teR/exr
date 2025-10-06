package webserver

import (
	"context"
	"net/http/httptest"
	"testing"

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

type mockExchangeRateService struct {
	addRatesFunc func(ctx context.Context) error
	getRatesFunc func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
}

func (m *mockExchangeRateService) GetRates(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
	if m.getRatesFunc != nil {
		return m.getRatesFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockExchangeRateService) AddRates(ctx context.Context) error {
	if m.addRatesFunc != nil {
		return m.addRatesFunc(ctx)
	}
	return nil
}

func TestNewServer(t *testing.T) {
	l := &mockLogger{}
	s := &mockExchangeRateService{}

	server := NewServer(l, s)

	if server == nil {
		t.Fatal("expected server instance")
	}
	if server.l != l {
		t.Error("logger should be assigned")
	}
	if server.service != s {
		t.Error("service should be assigned")
	}
	if server.refresher == nil {
		t.Error("refresher should be initialised")
	}
}

func TestServer_GetRouterInitialisesRoutes(t *testing.T) {
	l := &mockLogger{}
	s := &mockExchangeRateService{}
	server := NewServer(l, s)

	router := server.GetRouter()
	if router == nil {
		t.Fatal("router should not be nil")
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
