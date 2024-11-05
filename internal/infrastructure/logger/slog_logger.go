package logger

import (
	"log/slog"
	"os"

	"github.com/Mi7teR/exr/internal/interface/logger"
)

type SlogLogger struct {
	logger *slog.Logger
}

func NewSlogLogger() *SlogLogger {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	return &SlogLogger{
		logger: l,
	}
}

func (l *SlogLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

func (l *SlogLogger) With(args ...interface{}) logger.Logger {
	return &SlogLogger{
		logger: l.logger.With(args...),
	}
}
