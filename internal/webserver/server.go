package webserver

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Mi7teR/exr/internal/application/logger"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/service/exrate"
	"github.com/Mi7teR/exr/internal/web"
)

// ExchangeRateService определяет интерфейс для работы с курсами валют.
type ExchangeRateService interface {
	GetRates(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
	AddRates(ctx context.Context) error
}

type Server struct {
	l      logger.Logger
	uc     ExchangeRateService
	router *chi.Mux
	httpS  *http.Server
	bgCtx  context.Context
	cancel context.CancelFunc
}

func NewServer(l logger.Logger, uc ExchangeRateService) *Server {
	return &Server{l: l, uc: uc}
}

const (
	defaultCurrency      = "usd"
	currencyUSD          = "usd"
	currencyEUR          = "eur"
	currencyRUB          = "rub"
	refreshInterval      = 30 * time.Minute
	refreshTimeout       = 25 * time.Second
	percentageMultiplier = 100.0
)

func (s *Server) createRouter() *chi.Mux {
	router := chi.NewRouter()

	// ЧПУ маршруты
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = web.IndexPage(defaultCurrency).Render(r.Context(), w)
	})
	router.Get("/c/{currency}", s.handleCurrencyPage)

	return router
}

func (s *Server) Start(addr string) error {
	s.router = s.createRouter()

	// Контекст для фонового обновления и graceful shutdown.
	s.bgCtx, s.cancel = context.WithCancel(context.Background())

	// Фоновое обновление через заданный интервал.
	go s.startBackgroundRefresh(s.bgCtx)

	s.httpS = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s.httpS.ListenAndServe()
}

// GetRouter возвращает настроенный роутер для тестирования.
func (s *Server) GetRouter() http.Handler {
	if s.router == nil {
		s.router = s.createRouter()
	}
	return s.router
}

// Shutdown выполняет graceful shutdown HTTP-сервера и фоновых задач.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.httpS != nil {
		return s.httpS.Shutdown(ctx)
	}
	return nil
}

func (s *Server) startBackgroundRefresh(ctx context.Context) {
	// Первичная загрузка.
	s.refreshOnce(ctx)
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.refreshOnce(ctx)
		}
	}
}

func (s *Server) refreshOnce(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, refreshTimeout)
	defer cancel()
	if err := s.uc.AddRates(ctx); err != nil {
		s.l.Warn("background refresh failed", "err", err)
	}
}

func (s *Server) handleCurrencyPage(w http.ResponseWriter, r *http.Request) {
	currency := chi.URLParam(r, "currency")
	if currency == "" {
		currency = defaultCurrency
	}

	// Если это HTMX-запрос, возвращаем только содержимое таба
	if r.Header.Get("HX-Request") == "true" || r.Header.Get("Hx-Request") == "true" {
		banks, err := s.gatherBanks(r.Context(), currency)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = web.TabContent(banks, currency).Render(r.Context(), w)
		return
	}

	// Иначе рендерим полную страницу
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = web.IndexPage(currency).Render(r.Context(), w)
}

func (s *Server) gatherBanks(ctx context.Context, currency string) ([]web.Bank, error) {
	filter := &exrate.ExchangeRateFilter{StartDate: time.Time{}, EndDate: time.Time{}}
	rates, err := s.uc.GetRates(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get rates: %w", err)
	}
	banksMap := map[string]*web.Bank{}
	for _, r := range rates {
		b, ok := banksMap[r.Source]
		if !ok {
			b = &web.Bank{Name: r.Source}
			banksMap[r.Source] = b
		}
		buyRate, sellRate := parseBuySell(r)
		updateBankRates(b, r, buyRate, sellRate)
	}

	out := make([]web.Bank, 0, len(banksMap))
	for _, b := range banksMap {
		if shouldIncludeBank(b, currency) {
			out = append(out, *b)
		}
	}

	// Сортируем по возрастанию курса покупки выбранной валюты.
	sort.Slice(out, func(i, j int) bool {
		switch currency {
		case currencyUSD:
			iBuy, jBuy := fallbackBuy(out[i].Rates.USD.Buy, out[i].Rates.USD.Sell), fallbackBuy(out[j].Rates.USD.Buy, out[j].Rates.USD.Sell)
			return iBuy < jBuy
		case currencyEUR:
			iBuy, jBuy := fallbackBuy(out[i].Rates.EUR.Buy, out[i].Rates.EUR.Sell), fallbackBuy(out[j].Rates.EUR.Buy, out[j].Rates.EUR.Sell)
			return iBuy < jBuy
		case currencyRUB:
			iBuy, jBuy := fallbackBuy(out[i].Rates.RUB.Buy, out[i].Rates.RUB.Sell), fallbackBuy(out[j].Rates.RUB.Buy, out[j].Rates.RUB.Sell)
			return iBuy < jBuy
		default:
			return false
		}
	})
	return out, nil
}

func parseBuySell(r *entity.ExchangeRate) (buyRate, sellRate float64) {
	if r.Buy != "" {
		_, _ = fmt.Sscan(r.Buy, &buyRate)
	}
	if r.Sell != "" {
		_, _ = fmt.Sscan(r.Sell, &sellRate)
	}
	return buyRate, sellRate
}

func updateBankRates(b *web.Bank, r *entity.ExchangeRate, buyRate, sellRate float64) {
	switch r.CurrencyCode {
	case "USD":
		if buyRate > 0 {
			b.Rates.USD.Buy = buyRate
			b.Rates.USD.BuyChange = (r.BuyChangePrev / buyRate) * percentageMultiplier
		}
		if sellRate > 0 {
			b.Rates.USD.Sell = sellRate
			b.Rates.USD.SellChange = (r.SellChangePrev / sellRate) * percentageMultiplier
		}
	case "EUR":
		if buyRate > 0 {
			b.Rates.EUR.Buy = buyRate
			b.Rates.EUR.BuyChange = (r.BuyChangePrev / buyRate) * percentageMultiplier
		}
		if sellRate > 0 {
			b.Rates.EUR.Sell = sellRate
			b.Rates.EUR.SellChange = (r.SellChangePrev / sellRate) * percentageMultiplier
		}
	case "RUB":
		if buyRate > 0 {
			b.Rates.RUB.Buy = buyRate
			b.Rates.RUB.BuyChange = (r.BuyChangePrev / buyRate) * percentageMultiplier
		}
		if sellRate > 0 {
			b.Rates.RUB.Sell = sellRate
			b.Rates.RUB.SellChange = (r.SellChangePrev / sellRate) * percentageMultiplier
		}
	}
}

func shouldIncludeBank(b *web.Bank, currency string) bool {
	switch currency {
	case currencyUSD:
		return !(b.Rates.USD.Buy == 0 && b.Rates.USD.Sell == 0)
	case currencyEUR:
		return !(b.Rates.EUR.Buy == 0 && b.Rates.EUR.Sell == 0)
	case currencyRUB:
		return !(b.Rates.RUB.Buy == 0 && b.Rates.RUB.Sell == 0)
	default:
		return false
	}
}

func fallbackBuy(buy, sell float64) float64 {
	if buy == 0 {
		return sell
	}
	return buy
}
