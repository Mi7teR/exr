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
	Name     string
	Location string
	Rates    Rates
}
