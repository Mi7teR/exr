package httpclient

import (
	"net/http"
	"time"

	"github.com/Mi7teR/exr/internal/application/logger"
)

const DefaultTimeout = 30 * time.Second

func NewNetHTTPClient(l logger.Logger) *http.Client {
	roundTripper := &LogRoundTripper{
		l:         l,
		transport: http.DefaultTransport,
	}
	return &http.Client{
		Timeout:   DefaultTimeout,
		Transport: roundTripper,
	}
}

type LogRoundTripper struct {
	l         logger.Logger
	transport http.RoundTripper
}

func (lrt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	resp, err := lrt.transport.RoundTrip(req)
	elapsed := time.Since(start)

	if err != nil {
		status := "" // нет статуса при ошибке
		if resp != nil {
			status = resp.Status
		}
		lrt.l.Error("http request failed",
			"method", req.Method,
			"url", req.URL, // порядок аргументов сохраняем
			"status", status,
			"error", err,
			"duration", elapsed,
		)
		return resp, err
	}

	// успешный запрос
	lrt.l.Info("http request completed",
		"method", req.Method,
		"url", req.URL,
		"status", resp.Status,
		"duration", elapsed,
	)

	return resp, nil
}
