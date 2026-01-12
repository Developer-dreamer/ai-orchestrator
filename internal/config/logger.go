package config

import (
	"ai-orchestrator/internal/infra/telemetry/tracing"
	"log/slog"
	"os"
)

func NewLogger(level slog.Level) *slog.Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(tracing.TraceHandler{Handler: baseHandler})
	slog.SetDefault(logger)

	return logger
}
