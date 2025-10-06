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
	if filter == nil {
		filter = &ExchangeRateFilter{}
	}

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

	if err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return []*entity.ExchangeRate{}, nil
		}
		return nil, err
	}
	return rates, nil
}

func (u *ExchangeRateUsecase) AddRates(ctx context.Context) error {
	g := new(errgroup.Group)
	for _, d := range u.drivers {
		driverCopy := d
		g.Go(func() error {
			return u.fetchAndStore(ctx, driverCopy)
		})
	}
	return g.Wait()
}

func (u *ExchangeRateUsecase) fetchAndStore(ctx context.Context, d Driver) error {
	rates, fetchErr := d.FetchRates(ctx)
	if fetchErr != nil {
		return fetchErr
	}
	for _, rate := range rates {
		if err := u.upsertIfChanged(ctx, rate); err != nil {
			return err
		}
	}
	return nil
}

func (u *ExchangeRateUsecase) upsertIfChanged(ctx context.Context, rate *entity.ExchangeRate) error {
	lastRate, err := u.repo.GetLatestExchangeRate(ctx, rate.CurrencyCode, rate.Source)
	if err != nil {
		if errors.Is(err, internalErrors.ErrNotFound) {
			return u.repo.AddExchangeRate(ctx, rate)
		}
		return err
	}
	if lastRate.Buy == rate.Buy && lastRate.Sell == rate.Sell {
		return nil
	}
	return u.repo.AddExchangeRate(ctx, rate)
}
