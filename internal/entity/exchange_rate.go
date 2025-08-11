package entity

import "time"

// ExchangeRate represents an exchange rate.
type ExchangeRate struct {
	CurrencyCode   string
	Buy            string
	Sell           string
	Source         string
	CreatedAt      time.Time
	BuyChangePrev  float64 // текущее Buy - предыдущее Buy (0 если предыдущего нет)
	SellChangePrev float64 // текущее Sell - предыдущее Sell (0 если предыдущего нет)
}
