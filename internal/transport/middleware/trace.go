package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("uber-trace-id")

		if traceID == "" {
			traceID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), "trace_id", traceID)

		w.Header().Set("uber-trace-id", traceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
