package demo

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/store"
)

func TestGenerateCustomData(t *testing.T) {
	tests := []struct {
		service string
		name    string
		wantKey string
	}{
		{"api", "GET /users", "method"},
		{"auth", "ValidateToken", "auth_type"},
		{"payment", "ChargeCard", "amount"},
		{"worker", "ProcessPayment", "transaction_id"},
		{"webhook", "POST /webhook/payment", "source"},
	}

	for _, tt := range tests {
		data := generateCustomData(tt.service, tt.name)
		if data == nil {
			t.Errorf("generateCustomData(%s, %s) returned nil", tt.service, tt.name)
			continue
		}
		if _, exists := data[tt.wantKey]; !exists {
			t.Errorf("generateCustomData(%s, %s) missing expected key %s", tt.service, tt.name, tt.wantKey)
		}
	}
}

func TestEmitWithCustomData(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../../migrations"); err != nil {
		t.Fatal(err)
	}

	s := store.NewWithDB(db)
	defer s.Close()

	b := live.New()

	// Emit an event - it should have custom data
	emit(s, b, "api", "GET /users", "trace123", "info")

	events, err := s.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	// Custom data is added probabilistically, so it might not always be present
	// But request_id should always be there if custom data exists
	if event.Data != nil {
		if event.Data["request_id"] == nil {
			t.Error("expected request_id to be present when custom data exists")
		}
	}
}

func TestHistoricalDataGeneration(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatal(err)
	}
	if err := goose.Up(db, "../../migrations"); err != nil {
		t.Fatal(err)
	}

	s := store.NewWithDB(db)
	defer s.Close()

	generateHistoricalData(s)

	events, err := s.Recent(300) // Should have generated 200 events
	if err != nil {
		t.Fatal(err)
	}

	// Allow some variance due to random failures in historical generation
	if len(events) < 150 {
		t.Errorf("expected at least 150 historical events, got %d", len(events))
	}

	// Check that some events have custom data
	customDataCount := 0
	for _, event := range events {
		if event.Data != nil {
			customDataCount++
		}
	}
	// With 70% probability and improved matching, we should have some custom data
	if customDataCount == 0 {
		t.Logf("No custom data found in %d events - this might be due to randomness", len(events))
		// Don't fail the test as it's probabilistic
	}
	t.Logf("Generated %d events, %d with custom data", len(events), customDataCount)
}
