package webserver

import (
	"context"
	"time"

	"github.com/Mi7teR/exr/internal/application/logger"
)

type refresher struct {
	log       logger.Logger
	refreshFn func(ctx context.Context) error
	interval  time.Duration
	timeout   time.Duration
}

func newRefresher(log logger.Logger, refreshFn func(ctx context.Context) error, interval, timeout time.Duration) *refresher {
	return &refresher{
		log:       log,
		refreshFn: refreshFn,
		interval:  interval,
		timeout:   timeout,
	}
}

func (r *refresher) run(ctx context.Context) {
	r.tick(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.tick(ctx)
		}
	}
}

func (r *refresher) tick(parent context.Context) {
	runCtx, cancel := context.WithTimeout(parent, r.timeout)
	defer cancel()

	if err := r.refreshFn(runCtx); err != nil {
		r.log.Warn("background refresh failed", "err", err)
	}
}
