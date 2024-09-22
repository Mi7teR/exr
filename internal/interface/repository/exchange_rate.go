package repository

import (
	"context"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// ExchangeRateRepository is an interface that defines the methods that a repository must implement.
type ExchangeRateRepository interface {
	// GetExchangeRates returns a list of exchange rates.
	GetExchangeRates(ctx context.Context, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	// GetExchangeRatesByCurrencyCode returns a list of exchange rates by currency code.
	GetExchangeRatesByCurrencyCode(
		ctx context.Context, currencyCode string, startDate, endDate time.Time,
	) ([]*entity.ExchangeRate, error)
	// GetExchangeRatesByCurrencyCodeAndSource returns a list of exchange rates by currency code and source.
	GetExchangeRatesByCurrencyCodeAndSource(
		ctx context.Context,
		currencyCode, source string,
		startDate, endDate time.Time,
	) ([]*entity.ExchangeRate, error)
	// GetExchangeRatesBySource returns a list of exchange rates by source.
	GetExchangeRatesBySource(
		ctx context.Context, source string, startDate, endDate time.Time,
	) ([]*entity.ExchangeRate, error)
	// AddExchangeRate adds an exchange rate.
	AddExchangeRate(ctx context.Context, exchangeRate *entity.ExchangeRate) error
}
