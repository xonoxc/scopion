package store

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/xonoxc/scopion/internal/model"
)

func TestAppend(t *testing.T) {
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

	s := &Store{db: db}
	defer s.Close()

	err = s.Append(model.Event{
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
