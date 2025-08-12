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

// ExchangeRateService определяет интерфейс для работы с курсами валют
type ExchangeRateService interface {
	GetRates(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
	AddRates(ctx context.Context) error
}

type Server struct {
	l      logger.Logger
	uc     ExchangeRateService
	router *chi.Mux
}

func NewServer(l logger.Logger, uc ExchangeRateService) *Server {
	return &Server{l: l, uc: uc}
}

func (s *Server) createRouter() *chi.Mux {
	router := chi.NewRouter()

	// ЧПУ маршруты
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = web.IndexPage("usd").Render(r.Context(), w)
	})
	router.Get("/c/{currency}", s.handleCurrencyPage)

	return router
}

func (s *Server) Start(addr string) error {
	s.router = s.createRouter()

	// Фоновое обновление каждые 30 минут
	go s.startBackgroundRefresh()

	return http.ListenAndServe(addr, s.router)
}

// GetRouter возвращает настроенный роутер для тестирования
func (s *Server) GetRouter() http.Handler {
	if s.router == nil {
		s.router = s.createRouter()
	}
	return s.router
}

func (s *Server) startBackgroundRefresh() {
	// первичная загрузка
	s.refreshOnce()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.refreshOnce()
	}
}

func (s *Server) refreshOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	if err := s.uc.AddRates(ctx); err != nil {
		s.l.Warn("background refresh failed", "err", err)
	}
}

func (s *Server) handleCurrencyPage(w http.ResponseWriter, r *http.Request) {
	currency := chi.URLParam(r, "currency")
	if currency == "" {
		currency = "usd"
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
			b = &web.Bank{Name: r.Source, Location: "KZ"}
			banksMap[r.Source] = b
		}

		// Парсим курсы покупки и продажи
		var buyRate, sellRate float64
		if r.Buy != "" {
			_, _ = fmt.Sscan(r.Buy, &buyRate)
		}
		if r.Sell != "" {
			_, _ = fmt.Sscan(r.Sell, &sellRate)
		}

		// Заполняем курсы по валютам
		switch r.CurrencyCode {
		case "USD":
			if buyRate > 0 {
				b.Rates.USD.Buy = buyRate
				// Вычисляем процентное изменение для покупки
				b.Rates.USD.BuyChange = (r.BuyChangePrev / buyRate) * 100
			}
			if sellRate > 0 {
				b.Rates.USD.Sell = sellRate
				// Вычисляем процентное изменение для продажи
				b.Rates.USD.SellChange = (r.SellChangePrev / sellRate) * 100
			}
		case "EUR":
			if buyRate > 0 {
				b.Rates.EUR.Buy = buyRate
				b.Rates.EUR.BuyChange = (r.BuyChangePrev / buyRate) * 100
			}
			if sellRate > 0 {
				b.Rates.EUR.Sell = sellRate
				b.Rates.EUR.SellChange = (r.SellChangePrev / sellRate) * 100
			}
		case "RUB":
			if buyRate > 0 {
				b.Rates.RUB.Buy = buyRate
				b.Rates.RUB.BuyChange = (r.BuyChangePrev / buyRate) * 100
			}
			if sellRate > 0 {
				b.Rates.RUB.Sell = sellRate
				b.Rates.RUB.SellChange = (r.SellChangePrev / sellRate) * 100
			}
		}
	}

	out := make([]web.Bank, 0, len(banksMap))
	for _, b := range banksMap {
		// Добавляем только банки, у которых есть хотя бы один курс для запрошенной валюты
		switch currency {
		case "usd":
			if b.Rates.USD.Buy == 0 && b.Rates.USD.Sell == 0 {
				continue
			}
		case "eur":
			if b.Rates.EUR.Buy == 0 && b.Rates.EUR.Sell == 0 {
				continue
			}
		case "rub":
			if b.Rates.RUB.Buy == 0 && b.Rates.RUB.Sell == 0 {
				continue
			}
		}
		out = append(out, *b)
	}

	// сортируем по возрастанию курса покупки выбранной валюты
	sort.Slice(out, func(i, j int) bool {
		switch currency {
		case "usd":
			// Сортируем по курсу покупки, если он есть, иначе по курсу продажи
			iBuy, jBuy := out[i].Rates.USD.Buy, out[j].Rates.USD.Buy
			if iBuy == 0 {
				iBuy = out[i].Rates.USD.Sell
			}
			if jBuy == 0 {
				jBuy = out[j].Rates.USD.Sell
			}
			return iBuy < jBuy
		case "eur":
			iBuy, jBuy := out[i].Rates.EUR.Buy, out[j].Rates.EUR.Buy
			if iBuy == 0 {
				iBuy = out[i].Rates.EUR.Sell
			}
			if jBuy == 0 {
				jBuy = out[j].Rates.EUR.Sell
			}
			return iBuy < jBuy
		case "rub":
			iBuy, jBuy := out[i].Rates.RUB.Buy, out[j].Rates.RUB.Buy
			if iBuy == 0 {
				iBuy = out[i].Rates.RUB.Sell
			}
			if jBuy == 0 {
				jBuy = out[j].Rates.RUB.Sell
			}
			return iBuy < jBuy
		default:
			return false
		}
	})
	return out, nil
}
