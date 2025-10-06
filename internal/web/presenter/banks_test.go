package presenter

import (
	"testing"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
)

func TestNormalizeCurrency(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"usd", true},
		{"USD", true},
		{"eur", true},
		{"rub", true},
		{"gbp", false},
	}

	for _, tc := range tests {
		_, err := NormalizeCurrency(tc.input)
		if tc.ok && err != nil {
			t.Fatalf("expected success for %s", tc.input)
		}
		if !tc.ok && err == nil {
			t.Fatalf("expected error for %s", tc.input)
		}
	}
}

func TestBuildBanks(t *testing.T) {
	now := time.Now().UTC()
	rates := []*entity.ExchangeRate{
		{
			CurrencyCode:   "USD",
			Buy:            "490.50",
			Sell:           "495.00",
			Source:         "Kaspi",
			CreatedAt:      now,
			BuyChangePrev:  1.5,
			SellChangePrev: 2.0,
		},
		{
			CurrencyCode:   "USD",
			Buy:            "488.00",
			Sell:           "493.50",
			Source:         "Halyk",
			CreatedAt:      now.Add(-time.Hour),
			BuyChangePrev:  -0.5,
			SellChangePrev: -1.0,
		},
		{
			CurrencyCode:   "EUR",
			Buy:            "520.00",
			Sell:           "525.50",
			Source:         "Kaspi",
			CreatedAt:      now,
			BuyChangePrev:  2.5,
			SellChangePrev: 3.0,
		},
	}

	view, err := BuildBanks(rates, CurrencyUSD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(view.Banks) != 2 {
		t.Fatalf("expected 2 banks, got %d", len(view.Banks))
	}
	if view.LastUpdate.Before(now.Add(-time.Second)) {
		t.Fatalf("last update should reflect newest timestamp")
	}

	if view.Banks[0].Rates.USD.Buy < view.Banks[1].Rates.USD.Buy {
		t.Error("banks should be sorted by USD buy desc")
	}
}
