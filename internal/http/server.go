package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/trustlot/trustlot/internal/db"
	"github.com/trustlot/trustlot/internal/explain"
	"github.com/trustlot/trustlot/internal/replay"
	"github.com/trustlot/trustlot/internal/trust"
)

// Server is the HTTP API server.
type Server struct {
	router    chi.Router
	srv       *http.Server
	store     *db.Store
	explainer explain.Service
	replayer  replay.Service
	truster   trust.Service
}

// NewServer creates a configured HTTP server.
// store may be nil for testing without a database.
func NewServer(addr string, store *db.Store) *Server {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	var explainer explain.Service
	var replayer replay.Service
	var truster trust.Service
	if store != nil {
		explainer = explain.NewService(store)
		replayer = replay.NewService(store)
		truster = trust.NewService(store)
	}

	s := &Server{
		router: r,
		srv: &http.Server{
			Addr:              addr,
			Handler:           r,
			ReadHeaderTimeout: 10 * time.Second,
		},
		store:     store,
		explainer: explainer,
		replayer:  replayer,
		truster:   truster,
	}

	s.routes()
	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	slog.Info("http server starting", "addr", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http: listen: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("http server shutting down")
	return s.srv.Shutdown(ctx)
}
