package exrate

import (
	"context"
	"errors"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
	internalErrors "github.com/Mi7teR/exr/internal/errors"
	"golang.org/x/sync/errgroup"
)

// Driver is an interface that defines the methods that a driver must implement.
type Driver interface {
	// FetchRates returns a list of exchange rates.
	FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error)
}

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
	// GetLatestExchangeRate returns the most recent exchange rate for currency+source.
	GetLatestExchangeRate(ctx context.Context, currencyCode, source string) (*entity.ExchangeRate, error)
}

// ExchangeRateUsecase represents the usecase for exchange rates.
type ExchangeRateUsecase struct {
	repo    ExchangeRateRepository
	drivers map[string]Driver
}

func NewExchangeRateUsecase(repo ExchangeRateRepository, drivers map[string]Driver) *ExchangeRateUsecase {
	return &ExchangeRateUsecase{
		repo:    repo,
		drivers: drivers,
	}
}

// GetRates returns a list of exchange rates.
func (u *ExchangeRateUsecase) GetRates(
	ctx context.Context,
	filter *ExchangeRateFilter,
) ([]*entity.ExchangeRate, error) {
	var rates []*entity.ExchangeRate
	var err error

	switch {
	case filter.CurrencyCode != "" && filter.Source != "":
		rates, err = u.repo.GetExchangeRatesByCurrencyCodeAndSource(
			ctx,
			filter.CurrencyCode,
			filter.Source,
			filter.StartDate,
			filter.EndDate,
		)
	case filter.CurrencyCode != "":
		rates, err = u.repo.GetExchangeRatesByCurrencyCode(
			ctx,
			filter.CurrencyCode,
			filter.StartDate,
			filter.EndDate,
		)
	case filter.Source != "":
		rates, err = u.repo.GetExchangeRatesBySource(
			ctx,
			filter.Source,
			filter.StartDate,
			filter.EndDate,
		)
	default:
		rates, err = u.repo.GetExchangeRates(
			ctx,
			filter.StartDate,
			filter.EndDate,
		)
	}

	return rates, err
}

func (u *ExchangeRateUsecase) AddRates(ctx context.Context) error {
	g := new(errgroup.Group)
	for _, driver := range u.drivers {
		driver := driver // захватываем переменную для замыкания
		g.Go(func() error {
			rates, err := driver.FetchRates(ctx)
			if err != nil {
				return err
			}

			for _, rate := range rates {
				// Проверяем последний курс для этой валюты и источника
				lastRate, err := u.repo.GetLatestExchangeRate(ctx, rate.CurrencyCode, rate.Source)
				if err != nil {
					// Если курс не найден (первый запуск), сохраняем новый
					if errors.Is(err, internalErrors.ErrNotFound) {
						err = u.repo.AddExchangeRate(ctx, rate)
						if err != nil {
							return err
						}
						continue
					}
					// Другие ошибки просто возвращаем
					return err
				}

				// Сравниваем курсы - если одинаковые, пропускаем сохранение
				if lastRate.Buy == rate.Buy && lastRate.Sell == rate.Sell {
					// Курсы не изменились, пропускаем
					continue
				}

				// Курсы изменились, сохраняем новый
				err = u.repo.AddExchangeRate(ctx, rate)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
