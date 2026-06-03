package ingest

import (
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/store"
	"github.com/xonoxc/scopion/internal/store/migrations"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
	"github.com/xonoxc/scopion/internal/store/sqlite"
)

func TestSpanHandler(t *testing.T) {
	as, b := setupSpanTest(t)

	handler := SpanHandler(as, b)

	payload := `{"trace_id":"trace123","name":"http_request","service":"api","start_time":"2026-05-06T10:00:00Z","end_time":"2026-05-06T10:00:01Z","status":"OK"}`
	req := httptest.NewRequest("POST", "/ingest-span", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	store := as.Snapshot().Store
	spans, err := store.GetTraceSpans("trace123")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.TraceID != "trace123" {
		t.Errorf("Expected trace_id trace123, got %s", span.TraceID)
	}
	if span.Name != "http_request" {
		t.Errorf("Expected name http_request, got %s", span.Name)
	}
	if span.Service != "api" {
		t.Errorf("Expected service api, got %s", span.Service)
	}
	if span.Status != "OK" {
		t.Errorf("Expected status OK, got %s", span.Status)
	}
	if span.SpanID == "" {
		t.Error("Expected server-generated span_id to be non-empty")
	}
	if _, err := uuid.Parse(span.SpanID); err != nil {
		t.Errorf("Expected span_id to be valid UUID, got %s", span.SpanID)
	}
}

func TestSpanHandlerGeneratesSpanID(t *testing.T) {
	as, b := setupSpanTest(t)

	handler := SpanHandler(as, b)

	payload := `{"trace_id":"trace456","name":"db_query","service":"db"}`
	req := httptest.NewRequest("POST", "/ingest-span", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	store := as.Snapshot().Store
	spans, err := store.GetTraceSpans("trace456")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}
	if spans[0].SpanID == "" {
		t.Error("Expected server-generated span_id")
	}
}

func TestSpanHandlerWithParentSpanID(t *testing.T) {
	as, b := setupSpanTest(t)

	handler := SpanHandler(as, b)

	payload := `{"trace_id":"trace789","span_id":"should-be-ignored","parent_span_id":"parent123","name":"child_span","service":"svc"}`
	req := httptest.NewRequest("POST", "/ingest-span", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	store := as.Snapshot().Store
	spans, err := store.GetTraceSpans("trace789")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.ParentSpanID == nil || *span.ParentSpanID != "parent123" {
		t.Errorf("Expected parent_span_id parent123, got %v", span.ParentSpanID)
	}
	if span.SpanID == "should-be-ignored" {
		t.Error("Client-provided span_id should be ignored")
	}
}

func TestSpanHandlerDefaultsTimestamps(t *testing.T) {
	as, b := setupSpanTest(t)

	handler := SpanHandler(as, b)

	payload := `{"trace_id":"trace999","name":"test","service":"svc"}`
	req := httptest.NewRequest("POST", "/ingest-span", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	before := time.Now()
	handler.ServeHTTP(w, req)
	after := time.Now()

	if w.Code != 202 {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	store := as.Snapshot().Store
	spans, err := store.GetTraceSpans("trace999")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("Expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.StartTime.Before(before) || span.StartTime.After(after) {
		t.Errorf("Expected start_time between %v and %v, got %v", before, after, span.StartTime)
	}
	if span.EndTime.Before(before) || span.EndTime.After(after) {
		t.Errorf("Expected end_time between %v and %v, got %v", before, after, span.EndTime)
	}
}

func setupTestDB(t *testing.T) *sql.DB {
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

	return db
}

func setupSpanTest(t *testing.T) (*appcontext.AtomicAppState, *live.Broadcaster) {
	t.Helper()
	db := setupTestDB(t)
	s := sqlite.NewWithDB(db)
	as := appcontext.NewAtomicAppState(s, store.SINGLE_PRIMARY)
	b := live.New()
	return as, b
}
