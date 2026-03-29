package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/trustlot/trustlot/internal/config"
	"github.com/trustlot/trustlot/internal/db"
	thttp "github.com/trustlot/trustlot/internal/http"
	"github.com/trustlot/trustlot/internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	shutdownTelemetry := telemetry.Init(ctx, "trustlot-api")
	defer shutdownTelemetry()

	slog.Info("starting trustlot api",
		"addr", cfg.HTTPAddr,
		"env", cfg.Environment,
	)

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		slog.Warn("starting without database — endpoints will return empty data")
	}
	if pool != nil {
		defer pool.Close()
	}

	var store *db.Store
	if pool != nil {
		store = db.NewStore(pool)
	}

	srv := thttp.NewServer(cfg.HTTPAddr, store)

	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
