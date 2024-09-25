package logger

import (
	"log/slog"
	"os"

	"github.com/Mi7teR/exr/internal/interface/logger"
)

type slogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger() logger.Logger {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	return &slogLogger{
		logger: l,
	}
}

func (l *slogLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *slogLogger) With(args ...interface{}) logger.Logger {
	return &slogLogger{
		logger: l.logger.With(args...),
	}
}
