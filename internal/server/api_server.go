package server

import (
	"ai-orchestrator/internal/config"
	promptHandler "ai-orchestrator/internal/handler/prompt"
	"ai-orchestrator/internal/infra/redis"
	promptService "ai-orchestrator/internal/service/prompt"
	"ai-orchestrator/internal/util"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupServer(cfg *config.Config, logger *slog.Logger) *http.Server {
	redisClient, err := config.ConnectToRedis(cfg)
	if err != nil {
		logger.Error("Failed to initiate redis. Server shutdown.", "error", err)
		os.Exit(1)
	}

	streamOptions := &redis.StreamConfig{
		MaxBacklog:   1000,
		UseDelApprox: true,
		ReadCount:    1,
		BlockTime:    5 * time.Second,
	}
	rds := redis.NewService(logger, redisClient, streamOptions)
	ps := promptService.NewService(logger, rds, cfg)
	ph := promptHandler.NewHandler(logger, ps)

	r := registerRoutes(ph)
	logger.Info("Starting server")

	return &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}
}

func registerRoutes(handler *promptHandler.Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/ask", handler.PostPrompt).Methods(http.MethodPost)
	r.HandleFunc("/health", healthCheck).Methods(http.MethodGet)

	return r
}

func healthCheck(rw http.ResponseWriter, _ *http.Request) {
	util.WriteJSONResponse(rw, http.StatusOK, nil)
}

func GracefulShutdown(server *http.Server, logger *slog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("starting HTTP server", "addr", server.Addr, "idle_timeout", server.IdleTimeout, "read_timeout", server.ReadTimeout, "write_timeout", server.WriteTimeout)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error occurred when starting server", "error", err)
		}
	}()

	sig := <-sigChan
	logger.Info("received terminate, graceful shutdown", "sig", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("error during server shutdown", "error", err, "addr", server.Addr)
	}
}
