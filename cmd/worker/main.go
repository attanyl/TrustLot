package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/trustlot/trustlot/internal/config"
	"github.com/trustlot/trustlot/internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	shutdownTelemetry := telemetry.Init(ctx, "trustlot-worker")
	defer shutdownTelemetry()

	slog.Info("starting trustlot worker",
		"env", cfg.Environment,
	)

	// The worker will poll for ingestion jobs, reconciliation tasks,
	// and replay runs. For now, it blocks until signaled.
	slog.Info("worker waiting for jobs")
	<-ctx.Done()
	slog.Info("worker shutting down")
}
