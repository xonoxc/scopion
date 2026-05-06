package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store/migrations"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
	"github.com/xonoxc/scopion/internal/store/sqlite"
)

func setupStoreTest(t *testing.T) *sqlite.SqliteStore {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	migrator := migrations.NewMigrator(migrations.GetAll())
	if err := migrator.Migrate(db, migrateable.SQLITE); err != nil {
		t.Fatal(err)
	}

	return sqlite.NewWithDB(db)
}

func TestAppend(t *testing.T) {
	s := setupStoreTest(t)

	err := s.Append(model.Event{
		ID:        "test",
		Timestamp: time.Now(),
		Service:   "test",
		Name:      "test",
		TraceID:   "test",
		Level:     "info",
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := s.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestSearchEvents(t *testing.T) {
	s := setupStoreTest(t)

	baseTime := time.Now()

	// Insert events with different services
	events := []model.Event{
		{
			ID:        "event1",
			Timestamp: baseTime,
			Service:   "auth",
			Name:      "login",
			TraceID:   "trace1",
			Level:     "info",
			Data:      map[string]any{"user": "alice"},
		},
		{
			ID:        "event2",
			Timestamp: baseTime.Add(time.Second),
			Service:   "api",
			Name:      "request",
			TraceID:   "trace2",
			Level:     "info",
			Data:      map[string]any{"method": "GET"},
		},
		{
			ID:        "event3",
			Timestamp: baseTime.Add(2 * time.Second),
			Service:   "worker",
			Name:      "process",
			TraceID:   "trace1",
			Level:     "error",
		},
	}

	for _, event := range events {
		if err := s.Append(event); err != nil {
			t.Fatal(err)
		}
	}

	// Search by service
	results, err := s.SearchEvents("auth", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for auth search, got %d", len(results))
	}
	if results[0].ID == "" {
		t.Errorf("expected non-empty ID for auth search result")
	}

	// Search by trace ID
	results, err = s.SearchEvents("trace1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for trace1 search, got %d", len(results))
	}

	// Search with no matches
	results, err = s.SearchEvents("nonexistent", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for nonexistent search, got %d", len(results))
	}
}

func TestAppendWithCustomData(t *testing.T) {
	s := setupStoreTest(t)

	customData := map[string]any{
		"user_id":    123,
		"action":     "login",
		"ip_address": "192.168.1.1",
		"metadata": map[string]any{
			"browser": "chrome",
			"os":      "linux",
		},
	}

	err := s.Append(model.Event{
		ID:        "test-custom",
		Timestamp: time.Now(),
		Service:   "auth",
		Name:      "user_login",
		TraceID:   "trace123",
		Level:     "info",
		Data:      customData,
	})
	if err != nil {
		t.Fatal(err)
	}

	events, err := s.Recent(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.ID != "test-custom" {
		t.Errorf("expected ID test-custom, got %s", event.ID)
	}
	if event.Data == nil {
		t.Fatal("expected custom data to be present")
	}
	if event.Data["user_id"] != float64(123) {
		t.Errorf("expected user_id 123, got %v", event.Data["user_id"])
	}
	if event.Data["action"] != "login" {
		t.Errorf("expected action login, got %v", event.Data["action"])
	}
	metadata, ok := event.Data["metadata"].(map[string]any)
	if !ok {
		t.Fatal("expected metadata to be a map")
	}
	if metadata["browser"] != "chrome" {
		t.Errorf("expected browser chrome, got %v", metadata["browser"])
	}
}

func TestGetEventsByTraceID(t *testing.T) {
	s := setupStoreTest(t)

	traceID := "trace123"
	baseTime := time.Now()

	// Insert multiple events for the same trace
	for i := range 3 {
		event := model.Event{
			ID:        fmt.Sprintf("event-%d", i),
			Timestamp: baseTime.Add(time.Duration(i) * time.Second),
			Service:   "test",
			Name:      fmt.Sprintf("operation-%d", i),
			TraceID:   traceID,
			Level:     "info",
			Data: map[string]any{
				"step":  i,
				"value": i * 10,
			},
		}
		if err := s.Append(event); err != nil {
			t.Fatal(err)
		}
	}

	if err := s.Append(model.Event{
		ID:        "other-event",
		Timestamp: baseTime,
		Service:   "test",
		Name:      "other-operation",
		TraceID:   "other-trace",
		Level:     "info",
	}); err != nil {
		t.Fatal(err)
	}

	events, err := s.GetEventsByTraceID(traceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	for i := range 3 {
		if events[i].ID != fmt.Sprintf("event-%d", i) {
			t.Errorf("expected event ID event-%d, got %s", i, events[i].ID)
		}
		if events[i].Data["step"] != float64(i) {
			t.Errorf("expected step %d, got %v", i, events[i].Data["step"])
		}
	}
}

func TestGetTraces(t *testing.T) {
	s := setupStoreTest(t)

	baseTime := time.Now()

	spans := []model.Span{
		{
			TraceID:   "trace1",
			Name:      "GET /users",
			Service:   "api",
			StartTime: baseTime,
			EndTime:   baseTime.Add(100 * time.Millisecond),
			Status:    "OK",
		},
		{
			TraceID:   "trace1",
			Name:      "db query",
			Service:   "api",
			StartTime: baseTime.Add(50 * time.Millisecond),
			EndTime:   baseTime.Add(100 * time.Millisecond),
			Status:    "OK",
		},
		{
			TraceID:   "trace2",
			Name:      "ProcessPayment",
			Service:   "worker",
			StartTime: baseTime.Add(200 * time.Millisecond),
			EndTime:   baseTime.Add(200 * time.Millisecond),
			Status:    "ERROR",
		},
	}

	for _, span := range spans {
		if err := s.AppendSpan(span); err != nil {
			t.Fatal(err)
		}
	}

	traces, err := s.GetTraces(10)
	if err != nil {
		t.Fatalf("GetTraces failed: %v", err)
	}
	if len(traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(traces))
	}

	trace2 := traces[0]

	switch true {
	case trace2.ID != "trace2":
		t.Errorf("expected trace ID trace2, got %s", trace2.ID)

	case trace2.Spans != 1:
		t.Errorf("expected 1 span, got %d", trace2.Spans)

	case trace2.HasError != true:
		t.Errorf("expected error, got %v", trace2.HasError)

	case trace2.Duration != 0:
		t.Errorf("expected duration 0ms, got %d", trace2.Duration)

	}

	trace1 := traces[1]

	switch true {
	case trace1.ID != "trace1":
		t.Errorf("expected trace ID trace1, got %s", trace1.ID)

	case trace1.Spans != 2:
		t.Errorf("expected 2 spans, got %d", trace1.Spans)

	case trace1.HasError != false:
		t.Errorf("expected no error, got %v", trace1.HasError)

	case trace1.Duration != 100:
		t.Errorf("expected duration 100ms, got %d", trace1.Duration)

	}
}
