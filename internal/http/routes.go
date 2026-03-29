package http

import "github.com/go-chi/chi/v5"

func (s *Server) routes() {
	s.router.Get("/healthz", s.handleHealth())

	s.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/exceptions", s.handleListExceptions())
		r.Get("/exceptions/{id}", s.handleGetException())
		r.Get("/exceptions/{id}/explanation", s.handleGetExplanation())
		r.Get("/recon-runs", s.handleListReconRuns())
		r.Get("/trust-scores", s.handleListTrustScores())
		r.Get("/lineage/{entityType}/{entityID}", s.handleTraceLineage())

		r.Get("/replay-cases", s.handleListReplayCases())
		r.Get("/replay-cases/{id}", s.handleGetReplayCase())
		r.Post("/replay-cases/from-exception/{id}", s.handleCreateReplayCaseFromException())
		r.Post("/replay-cases/{id}/run", s.handleRunReplayCase())
	})
}
