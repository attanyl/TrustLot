package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpointNoStore(t *testing.T) {
	srv := NewServer(":0", nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", body["status"])
	}
	if body["database"] != "disconnected" {
		t.Errorf("expected database 'disconnected', got %q", body["database"])
	}
}

func TestExceptionsEndpointReturnsEmptyWithoutStore(t *testing.T) {
	srv := NewServer(":0", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/exceptions", nil)
	rec := httptest.NewRecorder()

	srv.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	exceptions, ok := body["exceptions"].([]any)
	if !ok {
		t.Fatal("expected exceptions array in response")
	}
	if len(exceptions) != 0 {
		t.Errorf("expected empty exceptions list, got %d", len(exceptions))
	}
}

func TestReconRunsEndpointReturnsEmptyWithoutStore(t *testing.T) {
	srv := NewServer(":0", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recon-runs", nil)
	rec := httptest.NewRecorder()

	srv.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	runs, ok := body["runs"].([]any)
	if !ok {
		t.Fatal("expected runs array in response")
	}
	if len(runs) != 0 {
		t.Errorf("expected empty runs list, got %d", len(runs))
	}
}

func TestCORSHeadersPresent(t *testing.T) {
	srv := NewServer(":0", nil)

	req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.router.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS Allow-Origin header")
	}
}
