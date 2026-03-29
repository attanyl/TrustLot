package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleHealth() http.HandlerFunc {
	type response struct {
		Status   string `json:"status"`
		Database string `json:"database"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		resp := response{Status: "ok", Database: "disconnected"}

		if s.store != nil {
			if err := s.store.Ping(r.Context()); err != nil {
				slog.Warn("health check: db ping failed", "error", err)
				resp.Database = "error"
			} else {
				resp.Database = "connected"
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func (s *Server) handleListExceptions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.store == nil {
			writeJSON(w, http.StatusOK, map[string]any{"exceptions": []any{}, "total": 0})
			return
		}

		status := r.URL.Query().Get("status")
		limit := queryInt(r, "limit", 100)

		exceptions, err := s.store.ListExceptions(r.Context(), status, limit)
		if err != nil {
			slog.Error("list exceptions", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"exceptions": exceptions,
			"total":      len(exceptions),
		})
	}
}

func (s *Server) handleGetException() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.store == nil {
			writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "not_implemented"})
			return
		}

		exc, err := s.store.GetException(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, exc)
	}
}

func (s *Server) handleGetExplanation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.explainer == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "explanation service unavailable"})
			return
		}

		expl, err := s.explainer.ExplainException(r.Context(), id)
		if err != nil {
			slog.Error("explain exception", "id", id, "error", err)
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, expl)
	}
}

func (s *Server) handleListReconRuns() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.store == nil {
			writeJSON(w, http.StatusOK, map[string]any{"runs": []any{}, "total": 0})
			return
		}

		limit := queryInt(r, "limit", 50)

		runs, err := s.store.ListReconRuns(r.Context(), limit)
		if err != nil {
			slog.Error("list recon runs", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"runs":  runs,
			"total": len(runs),
		})
	}
}

func (s *Server) handleListTrustScores() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.truster == nil {
			writeJSON(w, http.StatusOK, map[string]any{"scores": []any{}, "total": 0})
			return
		}

		limit := queryInt(r, "limit", 100)
		scores, err := s.truster.ListScores(r.Context(), limit)
		if err != nil {
			slog.Error("list trust scores", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"scores": scores,
			"total":  len(scores),
		})
	}
}

func (s *Server) handleTrustScoreForReconResult() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.truster == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "trust service unavailable"})
			return
		}

		result, err := s.truster.ComputeForReconResult(r.Context(), id)
		if err != nil {
			slog.Error("compute trust score for recon result", "id", id, "error", err)
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) handleTrustScoreForException() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.truster == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "trust service unavailable"})
			return
		}

		result, err := s.truster.ComputeForException(r.Context(), id)
		if err != nil {
			slog.Error("compute trust score for exception", "id", id, "error", err)
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func (s *Server) handleLineageForEntity() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityType := chi.URLParam(r, "entityType")
		entityID := chi.URLParam(r, "id")

		if s.liner == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "lineage service unavailable"})
			return
		}

		graph, err := s.liner.GetLineageForEntity(r.Context(), entityType, entityID)
		if err != nil {
			slog.Error("lineage for entity", "entity_type", entityType, "entity_id", entityID, "error", err)
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, graph)
	}
}

func (s *Server) handleLineageForException() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.liner == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "lineage service unavailable"})
			return
		}

		graph, err := s.liner.GetLineageForException(r.Context(), id)
		if err != nil {
			slog.Error("lineage for exception", "exception_id", id, "error", err)
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, graph)
	}
}

func (s *Server) handleListReplayCases() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.replayer == nil {
			writeJSON(w, http.StatusOK, map[string]any{"cases": []any{}, "total": 0})
			return
		}

		cases, err := s.replayer.ListCases(r.Context())
		if err != nil {
			slog.Error("list replay cases", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"cases": cases,
			"total": len(cases),
		})
	}
}

func (s *Server) handleGetReplayCase() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.replayer == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "replay service unavailable"})
			return
		}

		detail, err := s.replayer.GetCase(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, detail)
	}
}

func (s *Server) handleCreateReplayCaseFromException() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exceptionID := chi.URLParam(r, "id")

		if s.replayer == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "replay service unavailable"})
			return
		}

		rc, err := s.replayer.CreateCaseFromException(r.Context(), exceptionID)
		if err != nil {
			slog.Error("create replay case from exception", "exception_id", exceptionID, "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, rc)
	}
}

func (s *Server) handleRunReplayCase() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if s.replayer == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "replay service unavailable"})
			return
		}

		result, err := s.replayer.RunCase(r.Context(), id)
		if err != nil {
			slog.Error("run replay case", "case_id", id, "error", err)
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
