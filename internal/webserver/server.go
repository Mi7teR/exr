package webserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Mi7teR/exr/internal/application/logger"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/service/exrate"
	"github.com/Mi7teR/exr/internal/web/handler"
)

// ExchangeRateService определяет интерфейс для работы с курсами валют.
type ExchangeRateService interface {
	GetRates(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
	AddRates(ctx context.Context) error
}

type Server struct {
	l         logger.Logger
	service   ExchangeRateService
	router    *chi.Mux
	httpS     *http.Server
	cancel    context.CancelFunc
	refresher *refresher
}

func NewServer(l logger.Logger, uc ExchangeRateService) *Server {
	return &Server{
		l:         l,
		service:   uc,
		refresher: newRefresher(l, uc.AddRates, refreshInterval, refreshTimeout),
	}
}

const (
	refreshInterval = 30 * time.Minute
	refreshTimeout  = 25 * time.Second
)

func (s *Server) createRouter() *chi.Mux {
	r := chi.NewRouter()
	h := handler.NewRatesHandler(s.l, s.service)
	h.RegisterRoutes(r)
	return r
}

func (s *Server) Start(addr string) error {
	s.router = s.createRouter()
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.refresher.run(ctx)

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

func (s *Server) GetRouter() http.Handler {
	if s.router == nil {
		s.router = s.createRouter()
	}
	return s.router
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.httpS != nil {
		return s.httpS.Shutdown(ctx)
	}
	return nil
}
