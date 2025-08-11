package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/Mi7teR/exr/internal/entity"
	internalErrors "github.com/Mi7teR/exr/internal/errors"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func TestSQLiteExchangeRateRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo, err := NewSQLiteExchangeRateRepository(db)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Now().UTC()
	// Historical chain for USD/Kaspi
	usdKaspi1 := &entity.ExchangeRate{CurrencyCode: "USD", Buy: "440", Sell: "445", Source: "Kaspi", CreatedAt: now.Add(-3 * time.Hour)}
	usdKaspi2 := &entity.ExchangeRate{CurrencyCode: "USD", Buy: "445", Sell: "450", Source: "Kaspi", CreatedAt: now.Add(-2 * time.Hour)}
	usdKaspi3 := &entity.ExchangeRate{CurrencyCode: "USD", Buy: "450", Sell: "455", Source: "Kaspi", CreatedAt: now.Add(-1 * time.Hour)}
	// Another source chain USD/NBRK
	usdNbrk1 := &entity.ExchangeRate{CurrencyCode: "USD", Buy: "441", Sell: "446", Source: "NBRK", CreatedAt: now.Add(-2 * time.Hour)}
	usdNbrk2 := &entity.ExchangeRate{CurrencyCode: "USD", Buy: "451", Sell: "456", Source: "NBRK", CreatedAt: now.Add(-30 * time.Minute)}
	// EUR chain
	eurKaspi1 := &entity.ExchangeRate{CurrencyCode: "EUR", Buy: "500", Sell: "505", Source: "Kaspi", CreatedAt: now.Add(-45 * time.Minute)}
	eurKaspi2 := &entity.ExchangeRate{CurrencyCode: "EUR", Buy: "501", Sell: "506", Source: "Kaspi", CreatedAt: now.Add(-15 * time.Minute)}

	for _, r := range []*entity.ExchangeRate{usdKaspi1, usdKaspi2, usdKaspi3, usdNbrk1, usdNbrk2, eurKaspi1, eurKaspi2} {
		if err = repo.AddExchangeRate(ctx, r); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	all, err := repo.GetExchangeRates(ctx, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("get all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 latest got %d", len(all))
	}
	for _, r := range all {
		if r.CurrencyCode == "USD" && r.Source == "Kaspi" {
			if r.BuyChangePrev != 5 {
				t.Fatalf("Kaspi USD BuyChangePrev expected 5 got %v", r.BuyChangePrev)
			}
			if r.SellChangePrev != 5 {
				t.Fatalf("Kaspi USD SellChangePrev expected 5 got %v", r.SellChangePrev)
			}
		}
		if r.CurrencyCode == "USD" && r.Source == "NBRK" {
			if r.BuyChangePrev != 10 {
				t.Fatalf("NBRK USD BuyChangePrev expected 10 got %v", r.BuyChangePrev)
			}
		}
		if r.CurrencyCode == "EUR" && r.Source == "Kaspi" {
			if r.BuyChangePrev != 1 {
				t.Fatalf("EUR Kaspi BuyChangePrev expected 1 got %v", r.BuyChangePrev)
			}
		}
	}

	usd, err := repo.GetExchangeRatesByCurrencyCode(ctx, "USD", time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("get usd: %v", err)
	}
	if len(usd) != 2 {
		t.Fatalf("expected 2 usd latest, got %d", len(usd))
	}

	if _, err = repo.GetExchangeRatesByCurrencyCode(ctx, "RUB", time.Time{}, time.Time{}); err == nil || err != internalErrors.ErrNotFound {
		t.Fatalf("expected ErrNotFound for RUB, got %v", err)
	}
}
