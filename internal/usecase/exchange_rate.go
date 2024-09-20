package usecase

import (
	"fmt"

	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/interface/driver"
	"github.com/Mi7teR/exr/internal/interface/repository"
)

type ExchangeRateUsecase interface {
	GetRates(filter *ExchangeRateFilter) ([]*entity.ExchangeRate, error)
	AddRates()
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

func (u *exhangeRateUsecase) GetRates(filter *ExchangeRateFilter) ([]*entity.ExchangeRate, error) {
	var rates []*entity.ExchangeRate
	var err error

	switch {
	case filter.CurrencyCode != "" && filter.Source != "":
		rates, err = u.repo.GetExchangeRatesByCurrencyCodeAndSource(
			filter.CurrencyCode,
			filter.Source,
			filter.StartDate,
			filter.EndDate,
		)
	case filter.CurrencyCode != "":
		rates, err = u.repo.GetExchangeRatesByCurrencyCode(
			filter.CurrencyCode,
			filter.StartDate,
			filter.EndDate,
		)
	case filter.Source != "":
		rates, err = u.repo.GetExchangeRatesBySource(
			filter.Source,
			filter.StartDate,
			filter.EndDate,
		)
	default:
		rates, err = u.repo.GetExchangeRates(
			filter.StartDate,
			filter.EndDate,
		)
	}

	return rates, err
}

func (u *exhangeRateUsecase) AddRates() {
	for _, driver := range u.drivers {
		rates, err := driver.FetchRates()
		if err != nil {
			fmt.Println(err)
		}

		for _, rate := range rates {
			err := u.repo.AddExchangeRate(rate)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
