package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Mi7teR/exr/internal/application/logger"
	"github.com/Mi7teR/exr/internal/entity"
	"github.com/Mi7teR/exr/internal/service/exrate"
	"github.com/Mi7teR/exr/internal/web"
	"github.com/Mi7teR/exr/internal/web/presenter"
)

type RatesService interface {
	GetRates(ctx context.Context, filter *exrate.ExchangeRateFilter) ([]*entity.ExchangeRate, error)
}

type RatesHandler struct {
	log             logger.Logger
	service         RatesService
	defaultCurrency string
}

func NewRatesHandler(log logger.Logger, service RatesService) *RatesHandler {
	return &RatesHandler{
		log:             log,
		service:         service,
		defaultCurrency: presenter.CurrencyUSD,
	}
}

func (h *RatesHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.handleIndex)
	r.Get("/c/{currency}", h.handleCurrency)
}

func (h *RatesHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	view, err := h.fetchView(r.Context(), h.defaultCurrency)
	if err != nil {
		h.log.Error("get rates for index failed", "err", err)
	}
	h.renderIndex(w, r, h.defaultCurrency, view)
}

func (h *RatesHandler) handleCurrency(w http.ResponseWriter, r *http.Request) {
	currencyParam := chi.URLParam(r, "currency")
	if currencyParam == "" {
		currencyParam = h.defaultCurrency
	}

	currency, err := presenter.NormalizeCurrency(currencyParam)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	view, err := h.fetchView(r.Context(), currency)
	if err != nil {
		h.log.Error("get rates failed", "err", err)
		http.Error(w, "failed to load rates", http.StatusInternalServerError)
		return
	}

	if isHTMX(r) {
		respondHTML(w)
		updateLastUpdatedHeader(w, view.LastUpdate)
		_ = web.TabContent(view.Banks, view.LastUpdate, currency).Render(r.Context(), w)
		return
	}

	h.renderIndex(w, r, currency, view)
}

func (h *RatesHandler) renderIndex(w http.ResponseWriter, r *http.Request, currency string, view presenter.BanksView) {
	respondHTML(w)
	_ = web.IndexPage(currency, formatISO(view.LastUpdate)).Render(r.Context(), w)
}

func (h *RatesHandler) fetchView(ctx context.Context, currency string) (presenter.BanksView, error) {
	rates, err := h.service.GetRates(ctx, &exrate.ExchangeRateFilter{})
	if err != nil {
		return presenter.BanksView{}, err
	}
	return presenter.BuildBanks(rates, currency)
}

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" || r.Header.Get("Hx-Request") == "true"
}

func respondHTML(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func updateLastUpdatedHeader(w http.ResponseWriter, ts time.Time) {
	if ts.IsZero() {
		return
	}
	w.Header().Set("X-Last-Updated", ts.UTC().Format(time.RFC3339))
}

func formatISO(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
