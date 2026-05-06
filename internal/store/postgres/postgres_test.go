package postgres

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store/migrations"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

func postgresDSN() string {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/scopion_test?sslmode=disable"
	}
	return dsn
}

func setupTestDB(t *testing.T) *PostgresStore {
	t.Helper()

	dsn := postgresDSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Skipf("Postgres not available: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("Postgres not available: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS spans, events, goose_db_version CASCADE")
	if err != nil {
		t.Fatalf("failed to clean tables: %v", err)
	}

	migrator := migrations.NewMigrator(migrations.GetAll())
	if err := migrator.Migrate(db, migrateable.POSTGRES); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	store := NewWithDB(db)
	t.Cleanup(func() { store.Close() })

	return store
}

func TestPostgresAppendSpan(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	span := model.Span{
		TraceID:   "trace-1",
		Name:      "test-span",
		Service:   "test-service",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Status:    "OK",
	}

	if err := s.AppendSpan(span); err != nil {
		t.Fatal(err)
	}

	spans, err := s.GetTraceSpans("trace-1")
	if err != nil {
		t.Fatal(err)
	}

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	got := spans[0]
	if got.TraceID != "trace-1" {
		t.Errorf("expected trace_id trace-1, got %s", got.TraceID)
	}
	if got.SpanID == "" {
		t.Error("expected server-generated span_id, got empty string")
	}
	if got.Name != "test-span" {
		t.Errorf("expected name test-span, got %s", got.Name)
	}
	if got.Status != "OK" {
		t.Errorf("expected status OK, got %s", got.Status)
	}
}

func TestPostgresAppendSpanWithParentSpanID(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	parentID := "parent-123"
	span := model.Span{
		TraceID:      "trace-2",
		Name:         "child-span",
		Service:      "test-service",
		ParentSpanID: &parentID,
		StartTime:    time.Now(),
		EndTime:      time.Now().Add(50 * time.Millisecond),
		Status:       "ERROR",
	}

	if err := s.AppendSpan(span); err != nil {
		t.Fatal(err)
	}

	spans, err := s.GetTraceSpans("trace-2")
	if err != nil {
		t.Fatal(err)
	}

	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	if spans[0].ParentSpanID == nil {
		t.Fatal("expected parent_span_id to be set")
	}
	if *spans[0].ParentSpanID != "parent-123" {
		t.Errorf("expected parent_span_id parent-123, got %s", *spans[0].ParentSpanID)
	}
}

func TestPostgresGetTraceSpans(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	baseTime := time.Now()

	spans := []model.Span{
		{
			TraceID:   "trace-3",
			Name:      "span-a",
			Service:   "svc-a",
			StartTime: baseTime,
			EndTime:   baseTime.Add(10 * time.Millisecond),
			Status:    "OK",
		},
		{
			TraceID:   "trace-3",
			Name:      "span-b",
			Service:   "svc-b",
			StartTime: baseTime.Add(20 * time.Millisecond),
			EndTime:   baseTime.Add(30 * time.Millisecond),
			Status:    "OK",
		},
		{
			TraceID:   "trace-3",
			Name:      "span-c",
			Service:   "svc-a",
			StartTime: baseTime.Add(40 * time.Millisecond),
			EndTime:   baseTime.Add(50 * time.Millisecond),
			Status:    "ERROR",
		},
	}

	for _, span := range spans {
		if err := s.AppendSpan(span); err != nil {
			t.Fatal(err)
		}
	}

	result, err := s.GetTraceSpans("trace-3")
	if err != nil {
		t.Fatal(err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 spans, got %d", len(result))
	}

	if result[0].Name != "span-a" {
		t.Errorf("expected first span span-a, got %s", result[0].Name)
	}
	if result[1].Name != "span-b" {
		t.Errorf("expected second span span-b, got %s", result[1].Name)
	}
	if result[2].Name != "span-c" {
		t.Errorf("expected third span span-c, got %s", result[2].Name)
	}
}

func TestPostgresGetTraceSpansEmpty(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	spans, err := s.GetTraceSpans("nonexistent")
	if err != nil {
		t.Fatal(err)
	}

	if len(spans) != 0 {
		t.Errorf("expected 0 spans, got %d", len(spans))
	}
}

func TestPostgresAppendWithData(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	data := map[string]any{"key": "value", "count": float64(42)}
	event := model.Event{
		ID:        "event-with-data",
		Timestamp: time.Now(),
		Level:     "INFO",
		Service:   "test-svc",
		Name:      "test-event",
		Data:      data,
	}

	if err := s.Append(event); err != nil {
		t.Fatal(err)
	}

	events, err := s.SearchEvents("", 100)
	if err != nil {
		t.Fatal(err)
	}

	var found *model.Event
	for i, e := range events {
		if e.ID == "event-with-data" {
			found = &events[i]
			break
		}
	}

	if found == nil {
		t.Fatal("appended event not found in search results")
	}

	if found.Data == nil {
		t.Fatal("expected event Data to be non-nil, got nil")
	}

	if found.Data["key"] != "value" {
		t.Errorf("expected Data[key] = value, got %v", found.Data["key"])
	}
	if found.Data["count"] != float64(42) {
		t.Errorf("expected Data[count] = 42, got %v", found.Data["count"])
	}
}

func TestPostgresAppendWithNilData(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	event := model.Event{
		ID:        "event-nil-data",
		Timestamp: time.Now(),
		Level:     "INFO",
		Service:   "test-svc",
		Name:      "test-event",
		Data:      nil,
	}

	if err := s.Append(event); err != nil {
		t.Fatal(err)
	}

	events, err := s.SearchEvents("", 100)
	if err != nil {
		t.Fatal(err)
	}

	var found *model.Event
	for i, e := range events {
		if e.ID == "event-nil-data" {
			found = &events[i]
			break
		}
	}

	if found == nil {
		t.Fatal("appended event not found in search results")
	}

	if found.Data != nil {
		t.Errorf("expected event Data to be nil, got %v", found.Data)
	}
}

func TestPostgresGetTracesFromSpans(t *testing.T) {
	t.Parallel()
	s := setupTestDB(t)

	baseTime := time.Now()

	spans := []model.Span{
		{
			TraceID:   "trace-a",
			Name:      "span-1",
			Service:   "api",
			StartTime: baseTime,
			EndTime:   baseTime.Add(100 * time.Millisecond),
			Status:    "OK",
		},
		{
			TraceID:   "trace-a",
			Name:      "span-2",
			Service:   "db",
			StartTime: baseTime.Add(10 * time.Millisecond),
			EndTime:   baseTime.Add(50 * time.Millisecond),
			Status:    "OK",
		},
		{
			TraceID:   "trace-b",
			Name:      "span-3",
			Service:   "worker",
			StartTime: baseTime.Add(200 * time.Millisecond),
			EndTime:   baseTime.Add(250 * time.Millisecond),
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
		t.Fatal(err)
	}

	if len(traces) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(traces))
	}

	traceB := traces[0]
	if traceB.ID != "trace-b" {
		t.Errorf("expected first trace trace-b, got %s", traceB.ID)
	}
	if traceB.Spans != 1 {
		t.Errorf("expected trace-b to have 1 span, got %d", traceB.Spans)
	}
	if !traceB.HasError {
		t.Error("expected trace-b to have error")
	}

	traceA := traces[1]
	if traceA.ID != "trace-a" {
		t.Errorf("expected second trace trace-a, got %s", traceA.ID)
	}
	if traceA.Spans != 2 {
		t.Errorf("expected trace-a to have 2 spans, got %d", traceA.Spans)
	}
	if traceA.HasError {
		t.Error("expected trace-a to not have error")
	}
}
