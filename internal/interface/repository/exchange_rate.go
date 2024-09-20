package repository

import (
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

// ExchangeRateRepository is an interface that defines the methods that a repository must implement.
type ExchangeRateRepository interface {
	// GetExchangeRates returns a list of exchange rates.
	GetExchangeRates(startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	// GetExchangeRatesByCurrencyCode returns a list of exchange rates by currency code.
	GetExchangeRatesByCurrencyCode(currencyCode string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	// GetExchangeRatesByCurrencyCodeAndSource returns a list of exchange rates by currency code and source.
	GetExchangeRatesByCurrencyCodeAndSource(
		currencyCode, source string,
		startDate, endDate time.Time,
	) ([]*entity.ExchangeRate, error)
	// GetExchangeRatesBySource returns a list of exchange rates by source.
	GetExchangeRatesBySource(source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error)
	// AddExchangeRate adds an exchange rate.
	AddExchangeRate(exchangeRate *entity.ExchangeRate) error
}
