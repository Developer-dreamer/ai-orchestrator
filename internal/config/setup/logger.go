package setup

import (
	"ai-orchestrator/internal/common/logger"
	"log/slog"
	"os"
)

func NewLogger(level slog.Level) *slog.Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	l := slog.New(logger.TraceHandler{Handler: baseHandler})
	slog.SetDefault(l)

	return l
}
