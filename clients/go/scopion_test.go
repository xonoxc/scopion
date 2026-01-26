package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIngestEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ingest" && r.Method == "POST" {
			w.WriteHeader(202)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.IngestEvent("info", "test", "event", nil, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestIngestEventWithTrace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ingest" && r.Method == "POST" {
			w.WriteHeader(202)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	traceID := "trace123"
	err := client.IngestEvent("error", "api", "timeout", &traceID, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestIngestEventWithCustomData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ingest" && r.Method == "POST" {
			w.WriteHeader(202)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	customData := map[string]any{
		"user_id":    123,
		"action":     "login",
		"ip_address": "192.168.1.1",
	}
	err := client.IngestEvent("info", "auth", "user_login", nil, customData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestGetEvents(t *testing.T) {
	events := []Event{
		{ID: "1", Timestamp: "2023-01-01", Level: "info", Service: "test", Name: "event"},
	}
	jsonData, _ := json.Marshal(events)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/events" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			w.Write(jsonData)
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GetEvents(10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 1 || result[0].ID != "1" {
		t.Errorf("Expected one event with ID 1, got %v", result)
	}
}

func TestSubscribeLive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/live" && r.Method == "GET" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(200)
			w.Write([]byte("data: {\"id\": \"1\"}\n\n"))
			// Close after sending
			w.(http.Flusher).Flush()
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ch, err := client.SubscribeLive()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	select {
	case event := <-ch:
		if event.ID != "1" {
			t.Errorf("Expected event ID 1, got %s", event.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive an event within timeout")
	}
}
