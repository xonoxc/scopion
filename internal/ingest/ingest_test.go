package ingest

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/store"
)

func TestHandler(t *testing.T) {
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
	b := live.New()

	handler := Handler(s, b)

	req := httptest.NewRequest("POST", "/ingest", strings.NewReader(`{"level":"info","service":"test","name":"event"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	events, err := s.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestHandlerWithCustomData(t *testing.T) {
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
	b := live.New()

	handler := Handler(s, b)

	customData := `{
		"level": "info",
		"service": "auth",
		"name": "user_login",
		"trace_id": "trace123",
		"data": {
			"user_id": 123,
			"action": "login",
			"ip_address": "192.168.1.1",
			"device": {
				"type": "mobile",
				"os": "ios"
			}
		}
	}`

	req := httptest.NewRequest("POST", "/ingest", strings.NewReader(customData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	events, err := s.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Service != "auth" {
		t.Errorf("Expected service auth, got %s", event.Service)
	}
	if event.Data == nil {
		t.Fatal("Expected custom data to be present")
	}
	if event.Data["user_id"] != 123.0 {
		t.Errorf("Expected user_id 123.0, got %v", event.Data["user_id"])
	}
	if event.Data["action"] != "login" {
		t.Errorf("Expected action login, got %v", event.Data["action"])
	}

	device, ok := event.Data["device"].(map[string]any)
	if !ok {
		t.Fatal("Expected device to be a map")
	}

	if device["type"] != "mobile" {
		t.Errorf("Expected device type mobile, got %v", device["type"])
	}
}
