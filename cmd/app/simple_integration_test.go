package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/infrastructure/repository/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

func TestSimpleIntegration_DatabaseOperations(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Close()

	// Тестируем что подключение работает
	err = db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Инициализируем репозиторий
	repo, err := sqlite.NewSQLiteExchangeRateRepository(db)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Проверяем что таблица создалась
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='exchange_rates'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}

	if count == 0 {
		t.Fatalf("Table exchange_rates was not created")
	}

	t.Logf("Table created successfully")

	// Тестируем добавление курса (используем время в прошлом UTC)
	now := time.Now().UTC().Add(-time.Hour) // час назад в UTC
	testRate := &entity.ExchangeRate{
		CurrencyCode: "USD",
		Buy:          "490.50",
		Sell:         "495.00",
		Source:       "TestBank",
		CreatedAt:    now,
	}

	err = repo.AddExchangeRate(context.Background(), testRate)
	if err != nil {
		t.Fatalf("Failed to add exchange rate: %v", err)
	}

	t.Logf("Rate added successfully")

	// Тестируем получение последнего курса
	latestRate, err := repo.GetLatestExchangeRate(context.Background(), "USD", "TestBank")
	if err != nil {
		t.Fatalf("Failed to get latest rate: %v", err)
	}

	if latestRate.Buy != testRate.Buy || latestRate.Sell != testRate.Sell {
		t.Errorf("Retrieved rate doesn't match. Expected Buy=%s, Sell=%s, got Buy=%s, Sell=%s",
			testRate.Buy, testRate.Sell, latestRate.Buy, latestRate.Sell)
	}

	t.Logf("Rate retrieved successfully: Buy=%s, Sell=%s", latestRate.Buy, latestRate.Sell)

	// Добавляем второй курс для тестирования GetExchangeRates
	secondRate := &entity.ExchangeRate{
		CurrencyCode: "USD",
		Buy:          "491.00",
		Sell:         "496.00",
		Source:       "TestBank",
		CreatedAt:    now.Add(time.Minute),
	}

	err = repo.AddExchangeRate(context.Background(), secondRate)
	if err != nil {
		t.Fatalf("Failed to add second exchange rate: %v", err)
	}

	// Проверим сколько записей в базе напрямую
	var recordCount int
	err = db.QueryRow("SELECT COUNT(*) FROM exchange_rates").Scan(&recordCount)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}
	t.Logf("Records in database: %d", recordCount)

	// Попробуем простой запрос
	rows, err := db.Query("SELECT currency_code, buy, sell, source FROM exchange_rates")
	if err != nil {
		t.Fatalf("Failed to query records: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var curr, buy, sell, source string
		err = rows.Scan(&curr, &buy, &sell, &source)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		t.Logf("Record: %s %s %s %s", curr, buy, sell, source)
	}

	// Попробуем простой запрос сначала
	simpleQ := `SELECT currency_code, buy, sell, source, created_at FROM exchange_rates WHERE created_at BETWEEN ? AND ? ORDER BY created_at DESC`

	startDate := time.Unix(0, 0).UTC()
	endDate := time.Now().UTC()
	t.Logf("Query date range: %v to %v", startDate, endDate)

	rows2, err := db.Query(simpleQ, startDate, endDate)
	if err != nil {
		t.Fatalf("Failed to execute manual query: %v", err)
	}
	defer rows2.Close()

	rowCount := 0
	for rows2.Next() {
		rowCount++
		var currCode, buy, sell, source string
		var createdAt time.Time
		err = rows2.Scan(&currCode, &buy, &sell, &source, &createdAt)
		if err != nil {
			t.Fatalf("Failed to scan manual query row: %v", err)
		}
		t.Logf("Simple query result: %s %s %s %s %v", currCode, buy, sell, source, createdAt)
	}

	t.Logf("Manual query returned %d rows", rowCount)

	// Тестируем получение всех курсов
	rates, err := repo.GetExchangeRates(context.Background(), time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("Failed to get all rates: %v", err)
	}

	if len(rates) != 1 {
		t.Errorf("Expected 1 rate (latest per source), got %d", len(rates))
	}

	// Проверяем что получили последний курс
	if rates[0].Buy != secondRate.Buy {
		t.Errorf("Expected latest rate Buy=%s, got %s", secondRate.Buy, rates[0].Buy)
	}

	t.Logf("All rates retrieved successfully, count: %d, latest buy rate: %s", len(rates), rates[0].Buy)
}
