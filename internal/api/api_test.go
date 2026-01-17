package api

import (
	"database/sql"
	"encoding/json"
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
