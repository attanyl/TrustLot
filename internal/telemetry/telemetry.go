package telemetry

import (
	"context"
	"log/slog"
	"os"
)

// Init configures structured logging for the service.
// Returns a shutdown function that should be called on graceful exit.
// When OpenTelemetry is wired in, this will also initialize
// trace and metric providers.
func Init(_ context.Context, serviceName string) func() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler).With("service", serviceName)
	slog.SetDefault(logger)

	return func() {
		slog.Info("telemetry shutdown", "service", serviceName)
	}
}
