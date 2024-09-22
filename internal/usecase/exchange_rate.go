package usecase

import (
	"context"

	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/interface/driver"
	"github.com/Mi7teR/exr/internal/interface/repository"
	"golang.org/x/sync/errgroup"
)

type ExchangeRateUsecase interface {
	GetRates(ctx context.Context, filter *ExchangeRateFilter) ([]*entity.ExchangeRate, error)
	AddRates(ctx context.Context) error
}

type exhangeRateUsecase struct {
	repo    repository.ExchangeRateRepository
	drivers map[string]driver.Driver
}

func NewExchangeRateUsecase(repo repository.ExchangeRateRepository) ExchangeRateUsecase {
	return &exhangeRateUsecase{
		repo: repo,
	}
}

func (u *exhangeRateUsecase) GetRates(ctx context.Context, filter *ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
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

func (u *exhangeRateUsecase) AddRates(ctx context.Context) error {
	g := new(errgroup.Group)
	for _, driver := range u.drivers {
		g.Go(func() error {
			rates, err := driver.FetchRates(ctx)
			if err != nil {
				return err
			}

			for _, rate := range rates {
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
