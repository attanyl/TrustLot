package db

import (
	"context"
	"os"
	"testing"
	"time"
)

// These tests require a running Postgres instance.
// Set DATABASE_URL to enable them. They are skipped by default.

func testPool(t *testing.T) *Store {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, url)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	t.Cleanup(pool.Close)
	return NewStore(pool)
}

func TestStorePing(t *testing.T) {
	store := testPool(t)
	ctx := context.Background()

	if err := store.Ping(ctx); err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

func TestStoreListExceptions(t *testing.T) {
	store := testPool(t)
	ctx := context.Background()

	// Should not error even if tables are empty or seeded
	exceptions, err := store.ListExceptions(ctx, "", 10)
	if err != nil {
		t.Fatalf("list exceptions failed: %v", err)
	}

	// If seed data exists, verify structure
	if len(exceptions) > 0 {
		exc := exceptions[0]
		if exc.ID == "" {
			t.Error("exception ID should not be empty")
		}
		if exc.ReasonCode == "" {
			t.Error("exception reason code should not be empty")
		}
	}
}

func TestStoreListExceptionsWithStatusFilter(t *testing.T) {
	store := testPool(t)
	ctx := context.Background()

	exceptions, err := store.ListExceptions(ctx, "open", 100)
	if err != nil {
		t.Fatalf("list exceptions with filter failed: %v", err)
	}

	for _, exc := range exceptions {
		if exc.Status != "open" {
			t.Errorf("expected status 'open', got %q", exc.Status)
		}
	}
}

func TestStoreListReconRuns(t *testing.T) {
	store := testPool(t)
	ctx := context.Background()

	runs, err := store.ListReconRuns(ctx, 10)
	if err != nil {
		t.Fatalf("list recon runs failed: %v", err)
	}

	if len(runs) > 0 {
		run := runs[0]
		if run.ID == "" {
			t.Error("run ID should not be empty")
		}
		if run.EntityType == "" {
			t.Error("run entity type should not be empty")
		}
	}
}
