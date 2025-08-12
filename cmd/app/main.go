package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/Mi7teR/exr/internal/driver"
	"github.com/Mi7teR/exr/internal/infrastructure/httpclient"
	infraLogger "github.com/Mi7teR/exr/internal/infrastructure/logger"
	"github.com/Mi7teR/exr/internal/infrastructure/repository/sqlite"
	"github.com/Mi7teR/exr/internal/service/exrate"
	"github.com/Mi7teR/exr/internal/webserver"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	l := infraLogger.NewSlogLogger()

	// DB
	dsn := getenv("EXR_SQLITE_DSN", "file:exr.db?_foreign_keys=on")
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo, err := sqlite.NewSQLiteExchangeRateRepository(db)
	if err != nil {
		log.Fatalf("migrate repo: %v", err)
	}

	// HTTP client
	cli := httpclient.NewNetHTTPClient(l)

	// Инициализируем драйверы
	drivers := map[string]exrate.Driver{
		"Kaspi":   driver.NewKaspi("https://guide.kaspi.kz/client/api/v2/intgr/currency/rate/aggregate", cli),
		"Halyk":   driver.NewHalyk("https://back.halykbank.kz/common/currency-history", cli),
		"Freedom": driver.NewFreedom("https://bankffin.kz/api/exchange-rates/getRates", cli),
		"RBK":     driver.NewRBK("https://backend.bankrbk.kz/api/v1/modules/exchange_rates/data", cli),
		"HomeKZ":  driver.NewHome("https://home.kz/api/public/getCurrency", cli),
		"NBRK":    driver.NewNBRK("https://nationalbank.kz/rss/rates_all.xml", cli),
	}

	// Usecase с драйверами
	uc := exrate.NewExchangeRateUsecase(repo, drivers)

	addr := getenv("EXR_HTTP_ADDR", ":8080")
	server := webserver.NewServer(l, uc)
	l.Info("starting server", "addr", addr)
	if err := server.Start(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
