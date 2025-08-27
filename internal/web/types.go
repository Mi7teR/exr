//nolint:cyclop // Generated templ components in this package inflate average complexity; business code here is simple.
package web

type CurrencyRate struct {
	Buy        float64
	Sell       float64
	BuyChange  float64
	SellChange float64
}

type Rates struct {
	USD CurrencyRate
	EUR CurrencyRate
	RUB CurrencyRate
}

type Bank struct {
	Name  string
	Rates Rates
}
