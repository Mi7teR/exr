package webserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Mi7teR/exr/internal/application/logger"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/service/exrate"
)

// mockLogger реализует интерфейс logger.Logger для тестов
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...any)  {}
func (m *mockLogger) Info(msg string, args ...any)   {}
func (m *mockLogger) Warn(msg string, args ...any)   {}
func (m *mockLogger) Error(msg string, args ...any)  {}
func (m *mockLogger) With(args ...any) logger.Logger { return m }

// mockExchangeRateService реализует интерфейс ExchangeRateService для тестов
type mockExchangeRateService struct {
	getRatesFunc func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
	addRatesFunc func(ctx context.Context) error
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
	logger := &mockLogger{}
	service := &mockExchangeRateService{}

	server := NewServer(logger, service)

	if server == nil {
		t.Error("Expected server to be created, got nil")
		return
	}
	if server.l != logger {
		t.Error("Expected logger to be set correctly")
	}
	if server.uc != service {
		t.Error("Expected usecase to be set correctly")
	}
}

func TestServer_HandleCurrencyPage_FullPage(t *testing.T) {
	// Подготавливаем mock данные
	mockRates := []*entity.ExchangeRate{
		{
			CurrencyCode:   "USD",
			Buy:            "490.50",
			Sell:           "495.00",
			Source:         "Kaspi",
			CreatedAt:      time.Now(),
			BuyChangePrev:  1.5,
			SellChangePrev: 2.0,
		},
		{
			CurrencyCode:   "EUR",
			Buy:            "520.00",
			Sell:           "525.50",
			Source:         "Halyk",
			CreatedAt:      time.Now(),
			BuyChangePrev:  -0.5,
			SellChangePrev: -1.0,
		},
	}

	logger := &mockLogger{}
	service := &mockExchangeRateService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return mockRates, nil
		},
	}

	server := NewServer(logger, service)
	server.setupRouter()

	// Тест GET /c/usd (полная страница)
	req, err := http.NewRequest("GET", "/c/usd", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/html; charset=utf-8")
	}

	// Проверяем что HTML содержит основные элементы
	body := rr.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty response body")
	}
	// В реальной ситуации можно добавить более детальные проверки HTML
}

func TestServer_HandleCurrencyPage_HTMXRequest(t *testing.T) {
	// Подготавливаем mock данные
	mockRates := []*entity.ExchangeRate{
		{
			CurrencyCode:   "USD",
			Buy:            "490.50",
			Sell:           "495.00",
			Source:         "Kaspi",
			CreatedAt:      time.Now(),
			BuyChangePrev:  1.5,
			SellChangePrev: 2.0,
		},
	}

	logger := &mockLogger{}
	service := &mockExchangeRateService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return mockRates, nil
		},
	}

	server := NewServer(logger, service)
	server.setupRouter()

	// Тест HTMX запроса
	req, err := http.NewRequest("GET", "/c/usd", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("HX-Request", "true")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/html; charset=utf-8")
	}
}

func TestServer_GatherBanks(t *testing.T) {
	// Тестовые данные
	mockRates := []*entity.ExchangeRate{
		{
			CurrencyCode:   "USD",
			Buy:            "490.50",
			Sell:           "495.00",
			Source:         "Kaspi",
			CreatedAt:      time.Now(),
			BuyChangePrev:  1.5,
			SellChangePrev: 2.0,
		},
		{
			CurrencyCode:   "USD",
			Buy:            "488.00",
			Sell:           "493.50",
			Source:         "Halyk",
			CreatedAt:      time.Now(),
			BuyChangePrev:  -0.5,
			SellChangePrev: -1.0,
		},
		{
			CurrencyCode:   "EUR",
			Buy:            "520.00",
			Sell:           "525.50",
			Source:         "Kaspi",
			CreatedAt:      time.Now(),
			BuyChangePrev:  2.5,
			SellChangePrev: 3.0,
		},
	}

	logger := &mockLogger{}
	service := &mockExchangeRateService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return mockRates, nil
		},
	}

	server := NewServer(logger, service)

	// Тест сбора банков для USD
	banks, err := server.gatherBanks(context.Background(), "usd")
	if err != nil {
		t.Fatalf("gatherBanks returned error: %v", err)
	}

	if len(banks) != 2 {
		t.Errorf("Expected 2 banks for USD, got %d", len(banks))
	}

	// Проверяем что банки отсортированы по курсу покупки
	if len(banks) >= 2 {
		if banks[0].Rates.USD.Buy > banks[1].Rates.USD.Buy {
			t.Error("Banks should be sorted by buy rate ascending")
		}
	}

	// Проверяем корректность данных первого банка
	if len(banks) > 0 {
		bank := banks[0]
		if bank.Name == "" {
			t.Error("Bank name should not be empty")
		}
		if bank.Location != "KZ" {
			t.Errorf("Expected bank location to be KZ, got %s", bank.Location)
		}
		if bank.Rates.USD.Buy <= 0 {
			t.Error("USD buy rate should be greater than 0")
		}
		if bank.Rates.USD.Sell <= 0 {
			t.Error("USD sell rate should be greater than 0")
		}
	}

	// Тест сбора банков для EUR (должен вернуть только банки с EUR курсами)
	banks, err = server.gatherBanks(context.Background(), "eur")
	if err != nil {
		t.Fatalf("gatherBanks returned error: %v", err)
	}

	if len(banks) != 1 {
		t.Errorf("Expected 1 bank for EUR, got %d", len(banks))
	}

	if len(banks) > 0 {
		if banks[0].Rates.EUR.Buy <= 0 || banks[0].Rates.EUR.Sell <= 0 {
			t.Error("EUR rates should be greater than 0")
		}
	}
}

// Вспомогательная функция для настройки роутера в тестах
func (s *Server) setupRouter() {
	if s.router != nil {
		return
	}
	s.router = s.createRouter()
}

func BenchmarkServer_GatherBanks(b *testing.B) {
	// Создаем тестовые данные
	mockRates := make([]*entity.ExchangeRate, 100)
	sources := []string{"Kaspi", "Halyk", "Freedom", "RBK", "HomeKZ", "NBRK"}
	currencies := []string{"USD", "EUR", "RUB"}

	for i := 0; i < 100; i++ {
		mockRates[i] = &entity.ExchangeRate{
			CurrencyCode:   currencies[i%len(currencies)],
			Buy:            "490.50",
			Sell:           "495.00",
			Source:         sources[i%len(sources)],
			CreatedAt:      time.Now(),
			BuyChangePrev:  1.5,
			SellChangePrev: 2.0,
		}
	}

	logger := &mockLogger{}
	service := &mockExchangeRateService{
		getRatesFunc: func(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
			return mockRates, nil
		},
	}

	server := NewServer(logger, service)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.gatherBanks(context.Background(), "usd")
		if err != nil {
			b.Fatal(err)
		}
	}
}
