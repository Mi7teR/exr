package sqlite

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
	internalErrors "github.com/Mi7teR/exr/internal/errors"
)

// SQLiteExchangeRateRepository implements ExchangeRateRepository using SQLite.
type SQLiteExchangeRateRepository struct {
	db *sql.DB
}

// NewSQLiteExchangeRateRepository creates repository and applies schema.
func NewSQLiteExchangeRateRepository(db *sql.DB) (*SQLiteExchangeRateRepository, error) {
	repo := &SQLiteExchangeRateRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteExchangeRateRepository) migrate() error {
	const schema = `CREATE TABLE IF NOT EXISTS exchange_rates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		currency_code TEXT NOT NULL,
		buy TEXT NOT NULL,
		sell TEXT NOT NULL,
		source TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_exchange_rates_created_at ON exchange_rates(created_at);
	CREATE INDEX IF NOT EXISTS idx_exchange_rates_currency_code ON exchange_rates(currency_code);
	CREATE INDEX IF NOT EXISTS idx_exchange_rates_source ON exchange_rates(source);
	CREATE INDEX IF NOT EXISTS idx_exchange_rates_currency_source_created ON exchange_rates(currency_code, source, created_at);`
	_, err := r.db.Exec(schema)
	return err
}

// AddExchangeRate stores a new exchange rate.
func (r *SQLiteExchangeRateRepository) AddExchangeRate(ctx context.Context, rate *entity.ExchangeRate) error {
	if rate.CreatedAt.IsZero() {
		rate.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO exchange_rates(currency_code, buy, sell, source, created_at) VALUES(?,?,?,?,?)`,
		rate.CurrencyCode, rate.Buy, rate.Sell, rate.Source, rate.CreatedAt,
	)
	return err
}

// GetExchangeRates returns latest rates per currency+source in range with prev change.
func (r *SQLiteExchangeRateRepository) GetExchangeRates(ctx context.Context, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	q := `WITH latest AS (
		SELECT id, currency_code, buy, sell, source, created_at,
		ROW_NUMBER() OVER (PARTITION BY currency_code, source ORDER BY created_at DESC) rn
		FROM exchange_rates
		WHERE created_at BETWEEN ? AND ?
	)
	SELECT l.currency_code, l.buy, l.sell, l.source, l.created_at,
		(
			SELECT p.buy FROM exchange_rates p
			WHERE p.currency_code = l.currency_code AND p.source = l.source AND p.created_at < l.created_at
			ORDER BY p.created_at DESC LIMIT 1
		) AS prev_buy,
		(
			SELECT p.sell FROM exchange_rates p
			WHERE p.currency_code = l.currency_code AND p.source = l.source AND p.created_at < l.created_at
			ORDER BY p.created_at DESC LIMIT 1
		) AS prev_sell
	FROM latest l WHERE l.rn = 1
	ORDER BY l.created_at DESC`
	return r.queryRatesWithPrev(ctx, q, normalizeStart(startDate), normalizeEnd(endDate))
}

// GetExchangeRatesByCurrencyCode returns latest rate per source for the currency in range with prev change.
func (r *SQLiteExchangeRateRepository) GetExchangeRatesByCurrencyCode(ctx context.Context, currencyCode string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	q := `WITH latest AS (
		SELECT id, currency_code, buy, sell, source, created_at,
		ROW_NUMBER() OVER (PARTITION BY source ORDER BY created_at DESC) rn
		FROM exchange_rates
		WHERE currency_code = ? AND created_at BETWEEN ? AND ?
	)
	SELECT l.currency_code, l.buy, l.sell, l.source, l.created_at,
		(
			SELECT p.buy FROM exchange_rates p
			WHERE p.currency_code = l.currency_code AND p.source = l.source AND p.created_at < l.created_at
			ORDER BY p.created_at DESC LIMIT 1
		) AS prev_buy,
		(
			SELECT p.sell FROM exchange_rates p
			WHERE p.currency_code = l.currency_code AND p.source = l.source AND p.created_at < l.created_at
			ORDER BY p.created_at DESC LIMIT 1
		) AS prev_sell
	FROM latest l WHERE l.rn = 1
	ORDER BY l.created_at DESC`
	return r.queryRatesWithPrev(ctx, q, currencyCode, normalizeStart(startDate), normalizeEnd(endDate))
}

// GetExchangeRatesByCurrencyCodeAndSource returns rates filtered by currency & source.
func (r *SQLiteExchangeRateRepository) GetExchangeRatesByCurrencyCodeAndSource(ctx context.Context, currencyCode, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	q := `SELECT currency_code, buy, sell, source, created_at FROM exchange_rates WHERE currency_code = ? AND source = ? AND created_at BETWEEN ? AND ? ORDER BY created_at DESC`
	return r.queryRates(ctx, q, currencyCode, source, normalizeStart(startDate), normalizeEnd(endDate))
}

// GetExchangeRatesBySource returns rates filtered by source.
func (r *SQLiteExchangeRateRepository) GetExchangeRatesBySource(ctx context.Context, source string, startDate, endDate time.Time) ([]*entity.ExchangeRate, error) {
	q := `SELECT currency_code, buy, sell, source, created_at FROM exchange_rates WHERE source = ? AND created_at BETWEEN ? AND ? ORDER BY created_at DESC`
	return r.queryRates(ctx, q, source, normalizeStart(startDate), normalizeEnd(endDate))
}

func (r *SQLiteExchangeRateRepository) queryRates(ctx context.Context, query string, args ...any) ([]*entity.ExchangeRate, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*entity.ExchangeRate
	for rows.Next() {
		var rate entity.ExchangeRate
		if err = rows.Scan(&rate.CurrencyCode, &rate.Buy, &rate.Sell, &rate.Source, &rate.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &rate)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, internalErrors.ErrNotFound
	}
	return out, nil
}

func (r *SQLiteExchangeRateRepository) queryRatesWithPrev(ctx context.Context, query string, args ...any) ([]*entity.ExchangeRate, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*entity.ExchangeRate
	for rows.Next() {
		var rate entity.ExchangeRate
		var prevBuy, prevSell sql.NullString
		if err = rows.Scan(&rate.CurrencyCode, &rate.Buy, &rate.Sell, &rate.Source, &rate.CreatedAt, &prevBuy, &prevSell); err != nil {
			return nil, err
		}
		// compute changes
		if prevBuy.Valid {
			if curr, errParseCurr := strconv.ParseFloat(rate.Buy, 64); errParseCurr == nil {
				if prev, errParsePrev := strconv.ParseFloat(prevBuy.String, 64); errParsePrev == nil {
					rate.BuyChangePrev = curr - prev
				}
			}
		}
		if prevSell.Valid {
			if curr, errParseCurr := strconv.ParseFloat(rate.Sell, 64); errParseCurr == nil {
				if prev, errParsePrev := strconv.ParseFloat(prevSell.String, 64); errParsePrev == nil {
					rate.SellChangePrev = curr - prev
				}
			}
		}
		out = append(out, &rate)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, internalErrors.ErrNotFound
	}
	return out, nil
}

func normalizeStart(t time.Time) time.Time {
	if t.IsZero() {
		return time.Unix(0, 0)
	}
	return t
}

func normalizeEnd(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}
