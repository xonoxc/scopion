package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store"
)

func TestEventsHandler(t *testing.T) {
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

	// Insert event
	ts := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err = db.Exec(`INSERT INTO events (id, timestamp, level, service, name, trace_id) VALUES (?, ?, ?, ?, ?, ?)`,
		"test-id", ts, "info", "test", "event", "trace123")
	if err != nil {
		t.Fatal(err)
	}

	handler := EventsHandler(s)

	req := httptest.NewRequest("GET", "/api/events", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var events []model.Event
	err = json.NewDecoder(w.Body).Decode(&events)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != "test-id" {
		t.Errorf("Expected 1 event with ID test-id, got %v", events)
	}
}

func TestTraceEventsHandler(t *testing.T) {
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

	// Insert multiple events for the same trace
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 3 {
		ts := baseTime.Add(time.Duration(i) * time.Second)
		dataJSON := fmt.Sprintf(`{"step": %d, "value": %d}`, i, i*10)
		_, err = db.Exec(`INSERT INTO events (id, timestamp, level, service, name, trace_id, data) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("event-%d", i), ts, "info", "test", fmt.Sprintf("operation-%d", i), "trace123", dataJSON)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Insert event for different trace
	_, err = db.Exec(`INSERT INTO events (id, timestamp, level, service, name, trace_id) VALUES (?, ?, ?, ?, ?, ?)`,
		"other-event", baseTime, "info", "test", "other-operation", "other-trace")
	if err != nil {
		t.Fatal(err)
	}

	handler := TraceEventsHandler(s)

	req := httptest.NewRequest("GET", "/api/trace-events?trace_id=trace123", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var events []model.Event
	err = json.NewDecoder(w.Body).Decode(&events)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// Check ordering and data
	for i := range 3 {
		if events[i].ID != fmt.Sprintf("event-%d", i) {
			t.Errorf("Expected event ID event-%d, got %s", i, events[i].ID)
		}
		if events[i].Data == nil {
			t.Fatalf("Expected custom data for event %d", i)
		}
		if events[i].Data["step"] != float64(i) {
			t.Errorf("Expected step %d, got %v", i, events[i].Data["step"])
		}
	}
}

func TestTraceEventsHandlerNoTraceID(t *testing.T) {
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

	handler := TraceEventsHandler(s)

	req := httptest.NewRequest("GET", "/api/trace-events", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTraceEventsHandlerNotFound(t *testing.T) {
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

	handler := TraceEventsHandler(s)

	req := httptest.NewRequest("GET", "/api/trace-events?trace_id=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var events []model.Event
	err = json.NewDecoder(w.Body).Decode(&events)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events))
	}
}
