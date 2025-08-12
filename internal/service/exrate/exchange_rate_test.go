package exrate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
	internalErrors "github.com/Mi7teR/exr/internal/errors"
)

// mockRepository реализует интерфейс ExchangeRateRepository для тестов
type mockRepository struct {
	getRatesFunc                        func(ctx context.Context, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	getRatesByCurrencyCodeFunc          func(ctx context.Context, currencyCode string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	getRatesByCurrencyCodeAndSourceFunc func(ctx context.Context, currencyCode, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	getRatesBySourceFunc                func(ctx context.Context, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	addExchangeRateFunc                 func(ctx context.Context, exchangeRate *entity.ExchangeRate) error
	getLatestExchangeRateFunc           func(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error)
}

func (m *mockRepository) GetExchangeRates(ctx context.Context, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	if m.getRatesFunc != nil {
		return m.getRatesFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func (m *mockRepository) GetExchangeRatesByCurrencyCode(ctx context.Context, currencyCode string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	if m.getRatesByCurrencyCodeFunc != nil {
		return m.getRatesByCurrencyCodeFunc(ctx, currencyCode, startDate, endDate)
	}
	return nil, nil
}

func (m *mockRepository) GetExchangeRatesByCurrencyCodeAndSource(ctx context.Context, currencyCode, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	if m.getRatesByCurrencyCodeAndSourceFunc != nil {
		return m.getRatesByCurrencyCodeAndSourceFunc(ctx, currencyCode, source, startDate, endDate)
	}
	return nil, nil
}

func (m *mockRepository) GetExchangeRatesBySource(ctx context.Context, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	if m.getRatesBySourceFunc != nil {
		return m.getRatesBySourceFunc(ctx, source, startDate, endDate)
	}
	return nil, nil
}

func (m *mockRepository) AddExchangeRate(ctx context.Context, exchangeRate *entity.ExchangeRate) error {
	if m.addExchangeRateFunc != nil {
		return m.addExchangeRateFunc(ctx, exchangeRate)
	}
	return nil
}

func (m *mockRepository) GetLatestExchangeRate(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error) {
	if m.getLatestExchangeRateFunc != nil {
		return m.getLatestExchangeRateFunc(ctx, currencyCode, source)
	}
	return nil, internalErrors.ErrNotFound
}

// mockDriver реализует интерфейс Driver для тестов
type mockDriver struct {
	fetchRatesFunc func(ctx context.Context) ([]*entity.ExchangeRate, error)
}

func (m *mockDriver) FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error) {
	if m.fetchRatesFunc != nil {
		return m.fetchRatesFunc(ctx)
	}
	return nil, nil
}

func TestNewExchangeRateUsecase(t *testing.T) {
	repo := &mockRepository{}
	drivers := map[string]Driver{
		"test": &mockDriver{},
	}

	uc := NewExchangeRateUsecase(repo, drivers)

	if uc == nil {
		t.Error("Expected usecase to be created, got nil")
	}
	if uc.repo != repo {
		t.Error("Expected repository to be set correctly")
	}
	if len(uc.drivers) != 1 {
		t.Error("Expected drivers to be set correctly")
	}
}

func TestExchangeRateUsecase_GetRates(t *testing.T) {
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

	tests := []struct {
		name      string
		filter    *ExchangeRateFilter
		setupMock func(*mockRepository)
		want      []*entity.ExchangeRate
		wantErr   bool
	}{
		{
			name: "Get all rates",
			filter: &ExchangeRateFilter{
				StartDate: time.Time{},
				EndDate:   time.Time{},
			},
			setupMock: func(repo *mockRepository) {
				repo.getRatesFunc = func(ctx context.Context, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
					return mockRates, nil
				}
			},
			want:    mockRates,
			wantErr: false,
		},
		{
			name: "Get rates by currency code",
			filter: &ExchangeRateFilter{
				CurrencyCode: "USD",
				StartDate:    time.Time{},
				EndDate:      time.Time{},
			},
			setupMock: func(repo *mockRepository) {
				repo.getRatesByCurrencyCodeFunc = func(ctx context.Context, currencyCode string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
					return mockRates, nil
				}
			},
			want:    mockRates,
			wantErr: false,
		},
		{
			name: "Get rates by source",
			filter: &ExchangeRateFilter{
				Source:    "Kaspi",
				StartDate: time.Time{},
				EndDate:   time.Time{},
			},
			setupMock: func(repo *mockRepository) {
				repo.getRatesBySourceFunc = func(ctx context.Context, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
					return mockRates, nil
				}
			},
			want:    mockRates,
			wantErr: false,
		},
		{
			name: "Get rates by currency code and source",
			filter: &ExchangeRateFilter{
				CurrencyCode: "USD",
				Source:       "Kaspi",
				StartDate:    time.Time{},
				EndDate:      time.Time{},
			},
			setupMock: func(repo *mockRepository) {
				repo.getRatesByCurrencyCodeAndSourceFunc = func(ctx context.Context, currencyCode, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
					return mockRates, nil
				}
			},
			want:    mockRates,
			wantErr: false,
		},
		{
			name: "Repository error",
			filter: &ExchangeRateFilter{
				StartDate: time.Time{},
				EndDate:   time.Time{},
			},
			setupMock: func(repo *mockRepository) {
				repo.getRatesFunc = func(ctx context.Context, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
					return nil, errors.New("database error")
				}
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			uc := NewExchangeRateUsecase(repo, map[string]Driver{})
			got, err := uc.GetRates(context.Background(), tt.filter)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetRates() got = %d rates, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestExchangeRateUsecase_AddRates(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*mockRepository)
		setupDriver func(*mockDriver)
		wantErr     bool
	}{
		{
			name: "Add new rates successfully (first time)",
			setupRepo: func(repo *mockRepository) {
				repo.getLatestExchangeRateFunc = func(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error) {
					return nil, internalErrors.ErrNotFound
				}
				repo.addExchangeRateFunc = func(ctx context.Context, exchangeRate *entity.ExchangeRate) error {
					return nil
				}
			},
			setupDriver: func(driver *mockDriver) {
				driver.fetchRatesFunc = func(ctx context.Context) ([]*entity.ExchangeRate, error) {
					return []*entity.ExchangeRate{
						{
							CurrencyCode: "USD",
							Buy:          "490.50",
							Sell:         "495.00",
							Source:       "TestBank",
							CreatedAt:    time.Now(),
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Skip unchanged rates",
			setupRepo: func(repo *mockRepository) {
				repo.getLatestExchangeRateFunc = func(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error) {
					return &entity.ExchangeRate{
						CurrencyCode: "USD",
						Buy:          "490.50",
						Sell:         "495.00",
						Source:       "TestBank",
						CreatedAt:    time.Now(),
					}, nil
				}
				// addExchangeRateFunc не должна вызываться
			},
			setupDriver: func(driver *mockDriver) {
				driver.fetchRatesFunc = func(ctx context.Context) ([]*entity.ExchangeRate, error) {
					return []*entity.ExchangeRate{
						{
							CurrencyCode: "USD",
							Buy:          "490.50", // Same as last
							Sell:         "495.00", // Same as last
							Source:       "TestBank",
							CreatedAt:    time.Now(),
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Add changed rates",
			setupRepo: func(repo *mockRepository) {
				repo.getLatestExchangeRateFunc = func(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error) {
					return &entity.ExchangeRate{
						CurrencyCode: "USD",
						Buy:          "490.50",
						Sell:         "495.00",
						Source:       "TestBank",
						CreatedAt:    time.Now(),
					}, nil
				}
				repo.addExchangeRateFunc = func(ctx context.Context, exchangeRate *entity.ExchangeRate) error {
					return nil
				}
			},
			setupDriver: func(driver *mockDriver) {
				driver.fetchRatesFunc = func(ctx context.Context) ([]*entity.ExchangeRate, error) {
					return []*entity.ExchangeRate{
						{
							CurrencyCode: "USD",
							Buy:          "491.00", // Changed
							Sell:         "496.00", // Changed
							Source:       "TestBank",
							CreatedAt:    time.Now(),
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "Driver fetch error",
			setupRepo: func(repo *mockRepository) {
				// Repository methods should not be called
			},
			setupDriver: func(driver *mockDriver) {
				driver.fetchRatesFunc = func(ctx context.Context) ([]*entity.ExchangeRate, error) {
					return nil, errors.New("network error")
				}
			},
			wantErr: true,
		},
		{
			name: "Repository add error",
			setupRepo: func(repo *mockRepository) {
				repo.getLatestExchangeRateFunc = func(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error) {
					return nil, internalErrors.ErrNotFound
				}
				repo.addExchangeRateFunc = func(ctx context.Context, exchangeRate *entity.ExchangeRate) error {
					return errors.New("database error")
				}
			},
			setupDriver: func(driver *mockDriver) {
				driver.fetchRatesFunc = func(ctx context.Context) ([]*entity.ExchangeRate, error) {
					return []*entity.ExchangeRate{
						{
							CurrencyCode: "USD",
							Buy:          "490.50",
							Sell:         "495.00",
							Source:       "TestBank",
							CreatedAt:    time.Now(),
						},
					}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			driver := &mockDriver{}
			drivers := map[string]Driver{"TestBank": driver}

			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			if tt.setupDriver != nil {
				tt.setupDriver(driver)
			}

			uc := NewExchangeRateUsecase(repo, drivers)
			err := uc.AddRates(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("AddRates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkExchangeRateUsecase_AddRates(b *testing.B) {
	repo := &mockRepository{
		getLatestExchangeRateFunc: func(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error) {
			return nil, internalErrors.ErrNotFound
		},
		addExchangeRateFunc: func(ctx context.Context, exchangeRate *entity.ExchangeRate) error {
			return nil
		},
	}

	driver := &mockDriver{
		fetchRatesFunc: func(ctx context.Context) ([]*entity.ExchangeRate, error) {
			// Simulate fetching 10 rates
			rates := make([]*entity.ExchangeRate, 10)
			for i := 0; i < 10; i++ {
				rates[i] = &entity.ExchangeRate{
					CurrencyCode: "USD",
					Buy:          "490.50",
					Sell:         "495.00",
					Source:       "TestBank",
					CreatedAt:    time.Now(),
				}
			}
			return rates, nil
		},
	}

	drivers := map[string]Driver{"TestBank": driver}
	uc := NewExchangeRateUsecase(repo, drivers)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := uc.AddRates(context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
}
