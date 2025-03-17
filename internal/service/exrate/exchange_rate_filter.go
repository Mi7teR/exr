package exrate

import "time"

// ExchangeRateFilter is a struct that contains the filter parameters for exchange rates.
type ExchangeRateFilter struct {
	CurrencyCode string
	Source       string
	StartDate    time.Time
	EndDate      time.Time
}
