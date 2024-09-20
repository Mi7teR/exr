package driver

import (
	"context"

	"github.com/Mi7teR/exr/internal/entity"
)

// Driver is an interface that defines the methods that a driver must implement.
type Driver interface {
	// FetchRates returns a list of exchange rates.
	FetchRates(ctx context.Context) ([]*entity.ExchangeRate, error)
}
