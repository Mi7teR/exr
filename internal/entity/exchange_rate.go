package entity

import "time"

// ExchangeRate represents an exchange rate.
type ExchangeRate struct {
	CurrencyCode string
	Buy          string
	Sell         string
	Source       string
	CreatedAt    time.Time
}
