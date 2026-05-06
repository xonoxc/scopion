package api

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store/migrations"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
	"github.com/xonoxc/scopion/internal/store/sqlite"
)

func setupTraceSpansTest(t *testing.T) *appcontext.AtomicAppState {
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

	s := sqlite.NewWithDB(db)

	spans := []struct {
		traceID string
		name    string
		start   time.Time
	}{
		{"trace1", "span-a", time.Now().Add(-2 * time.Hour)},
		{"trace1", "span-b", time.Now().Add(-1 * time.Hour)},
		{"trace1", "span-c", time.Now()},
		{"trace2", "span-x", time.Now().Add(-1 * time.Hour)},
	}

	for _, sp := range spans {
		span := model.Span{
			TraceID:   sp.traceID,
			Name:      sp.name,
			Service:   "test-svc",
			StartTime: sp.start,
			EndTime:   sp.start.Add(100 * time.Millisecond),
			Status:    "OK",
		}
		if err := s.AppendSpan(span); err != nil {
			t.Fatal(err)
		}
	}

	as := appcontext.NewAtomicAppState(s, "sqlite")
	return as
}

func TestTraceSpansHandler(t *testing.T) {
	as := setupTraceSpansTest(t)

	handler := TraceSpansHandler(as)

	req := httptest.NewRequest("GET", "/api/trace-spans?trace_id=trace1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var spans []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&spans); err != nil {
		t.Fatal(err)
	}

	if len(spans) != 3 {
		t.Fatalf("Expected 3 spans, got %d", len(spans))
	}

	names := []string{spans[0]["name"].(string), spans[1]["name"].(string), spans[2]["name"].(string)}
	if names[0] != "span-a" || names[1] != "span-b" || names[2] != "span-c" {
		t.Errorf("Expected spans ordered by start_time, got %v", names)
	}
}

func TestTraceSpansHandlerMissingTraceID(t *testing.T) {
	as := setupTraceSpansTest(t)

	handler := TraceSpansHandler(as)

	req := httptest.NewRequest("GET", "/api/trace-spans", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestTraceSpansHandlerEmptyTrace(t *testing.T) {
	as := setupTraceSpansTest(t)

	handler := TraceSpansHandler(as)

	req := httptest.NewRequest("GET", "/api/trace-spans?trace_id=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var spans []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&spans); err != nil {
		t.Fatal(err)
	}

	if len(spans) != 0 {
		t.Errorf("Expected 0 spans, got %d", len(spans))
	}
}
