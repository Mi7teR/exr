package presenter

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/web"
)

const (
	CurrencyUSD = "usd"
	CurrencyEUR = "eur"
	CurrencyRUB = "rub"
)

var ErrUnsupportedCurrency = errors.New("unsupported currency")

type BanksView struct {
	Banks      []web.Bank
	LastUpdate time.Time
}

func NormalizeCurrency(value string) (string, error) {
	switch strings.ToLower(value) {
	case CurrencyUSD:
		return CurrencyUSD, nil
	case CurrencyEUR:
		return CurrencyEUR, nil
	case CurrencyRUB:
		return CurrencyRUB, nil
	default:
		return "", ErrUnsupportedCurrency
	}
}

func BuildBanks(rates []*entity.ExchangeRate, currency string) (BanksView, error) {
	if _, err := NormalizeCurrency(currency); err != nil {
		return BanksView{}, err
	}

	banksMap := make(map[string]*web.Bank)
	var latest time.Time

	for _, rate := range rates {
		if rate == nil {
			continue
		}
		if rate.CreatedAt.After(latest) {
			latest = rate.CreatedAt
		}
		bank := ensureBank(banksMap, rate.Source)
		buy, sell := parseRate(rate.Buy), parseRate(rate.Sell)
		updateBankRates(bank, rate.CurrencyCode, buy, sell, rate.BuyChangePrev, rate.SellChangePrev)
	}

	banks := filterBanks(banksMap, currency)
	sortBanks(banks, currency)

	if latest.IsZero() {
		latest = time.Now().UTC()
	}

	return BanksView{Banks: banks, LastUpdate: latest}, nil
}

func ensureBank(store map[string]*web.Bank, source string) *web.Bank {
	if source == "" {
		source = "unknown"
	}
	if b, ok := store[source]; ok {
		return b
	}
	bank := &web.Bank{Name: source}
	store[source] = bank
	return bank
}

func parseRate(value string) float64 {
	if value == "" {
		return 0
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return f
}

func updateBankRates(bank *web.Bank, currencyCode string, buy, sell, buyDelta, sellDelta float64) {
	switch strings.ToUpper(currencyCode) {
	case "USD":
		assignCurrency(&bank.Rates.USD, buy, sell, buyDelta, sellDelta)
	case "EUR":
		assignCurrency(&bank.Rates.EUR, buy, sell, buyDelta, sellDelta)
	case "RUB":
		assignCurrency(&bank.Rates.RUB, buy, sell, buyDelta, sellDelta)
	}
}

func assignCurrency(rate *web.CurrencyRate, buy, sell, buyDelta, sellDelta float64) {
	if buy > 0 {
		rate.Buy = buy
		rate.BuyChange = percentageChange(buy, buyDelta)
	}
	if sell > 0 {
		rate.Sell = sell
		rate.SellChange = percentageChange(sell, sellDelta)
	}
}

func percentageChange(current, delta float64) float64 {
	if delta == 0 {
		return 0
	}
	prev := current - delta
	if prev <= 0 {
		return 0
	}
	return (delta / prev) * 100
}

func filterBanks(banks map[string]*web.Bank, currency string) []web.Bank {
	result := make([]web.Bank, 0, len(banks))
	for _, bank := range banks {
		if bankHasCurrency(bank, currency) {
			result = append(result, *bank)
		}
	}
	return result
}

func bankHasCurrency(bank *web.Bank, currency string) bool {
	switch currency {
	case CurrencyUSD:
		return !(bank.Rates.USD.Buy == 0 && bank.Rates.USD.Sell == 0)
	case CurrencyEUR:
		return !(bank.Rates.EUR.Buy == 0 && bank.Rates.EUR.Sell == 0)
	case CurrencyRUB:
		return !(bank.Rates.RUB.Buy == 0 && bank.Rates.RUB.Sell == 0)
	default:
		return false
	}
}

func sortBanks(banks []web.Bank, currency string) {
	sort.Slice(banks, func(i, j int) bool {
		left := fallbackBuy(banks[i], currency)
		right := fallbackBuy(banks[j], currency)
		if left == right {
			return banks[i].Name < banks[j].Name
		}
		return left > right
	})
}

func fallbackBuy(bank web.Bank, currency string) float64 {
	switch currency {
	case CurrencyUSD:
		if bank.Rates.USD.Buy > 0 {
			return bank.Rates.USD.Buy
		}
		return bank.Rates.USD.Sell
	case CurrencyEUR:
		if bank.Rates.EUR.Buy > 0 {
			return bank.Rates.EUR.Buy
		}
		return bank.Rates.EUR.Sell
	case CurrencyRUB:
		if bank.Rates.RUB.Buy > 0 {
			return bank.Rates.RUB.Buy
		}
		return bank.Rates.RUB.Sell
	default:
		return 0
	}
}
