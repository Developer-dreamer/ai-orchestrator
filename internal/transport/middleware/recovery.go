package middleware

import (
	"ai-orchestrator/internal/common/logger"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type RecoveryManager struct {
	logger logger.Logger
}

func NewRecoveryManager(logger logger.Logger) *RecoveryManager {
	return &RecoveryManager{
		logger: logger,
	}
}

func (rm *RecoveryManager) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				if err == http.ErrAbortHandler {
					panic(err)
				}

				stack := debug.Stack()

				rm.logger.ErrorContext(r.Context(), "PANIC RECOVERED",
					slog.Any("error", err),
					slog.String("stack", string(stack)),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal Server Error"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
