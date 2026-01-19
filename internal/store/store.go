package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xonoxc/scopion/internal/model"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{db: db}, nil
}

func NewWithDB(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Append(e model.Event) error {
	var dataJSON []byte
	var err error
	if e.Data != nil {
		dataJSON, err = json.Marshal(e.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}
	}

	_, err = s.db.Exec(
		"INSERT INTO events (id, timestamp, level, service, name, trace_id, data) VALUES (?, ?, ?, ?, ?, ?, ?)",
		e.ID, e.Timestamp, e.Level, e.Service, e.Name, e.TraceID, string(dataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

func (s *Store) Recent(n int) ([]model.Event, error) {
	rows, err := s.db.Query(
		"SELECT id, timestamp, level, service, name, trace_id, data FROM events ORDER BY timestamp DESC LIMIT ?",
		n,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		var dataStr sql.NullString
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Service, &e.Name, &e.TraceID, &dataStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if dataStr.Valid && dataStr.String != "" {
			err = json.Unmarshal([]byte(dataStr.String), &e.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
			}
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events, nil
}

type Stats struct {
	TotalEvents    int     `json:"total_events"`
	ErrorRate      float64 `json:"error_rate"`
	ActiveServices int     `json:"active_services"`
}

func (s *Store) GetStats() (*Stats, error) {
	var totalEvents int
	err := s.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&totalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get total events: %w", err)
	}

	var errorEvents int
	err = s.db.QueryRow("SELECT COUNT(*) FROM events WHERE level = 'error'").Scan(&errorEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to get error events: %w", err)
	}

	var activeServices int
	err = s.db.QueryRow("SELECT COUNT(DISTINCT service) FROM events").Scan(&activeServices)
	if err != nil {
		return nil, fmt.Errorf("failed to get active services: %w", err)
	}

	var errorRate float64
	if totalEvents > 0 {
		errorRate = float64(errorEvents) / float64(totalEvents) * 100
	}

	return &Stats{
		TotalEvents:    totalEvents,
		ErrorRate:      errorRate,
		ActiveServices: activeServices,
	}, nil
}

type ErrorByService struct {
	Service string `json:"service"`
	Count   int    `json:"count"`
}

func (s *Store) GetErrorsByService(hours int) ([]ErrorByService, error) {
	query := `
		SELECT service, COUNT(*) as count
		FROM events
		WHERE level = 'error' AND timestamp >= datetime('now', '-%d hours')
		GROUP BY service
		ORDER BY count DESC
	`
	query = fmt.Sprintf(query, hours)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query errors by service: %w", err)
	}
	defer rows.Close()

	var results []ErrorByService
	for rows.Next() {
		var e ErrorByService
		err := rows.Scan(&e.Service, &e.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan error by service: %w", err)
		}
		results = append(results, e)
	}

	return results, rows.Err()
}

type ServiceInfo struct {
	Name         string    `json:"name"`
	ErrorCount   int       `json:"error_count"`
	LastActivity time.Time `json:"last_activity"`
	EventCount   int       `json:"event_count"`
}

func (s *Store) GetServices() ([]ServiceInfo, error) {
	query := `
		SELECT
			service,
			COUNT(CASE WHEN level = 'error' THEN 1 END) as error_count,
			MAX(timestamp) as last_activity,
			COUNT(*) as event_count
		FROM events
		GROUP BY service
		ORDER BY last_activity DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var results []ServiceInfo
	for rows.Next() {
		var s ServiceInfo
		err := rows.Scan(&s.Name, &s.ErrorCount, &s.LastActivity, &s.EventCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service info: %w", err)
		}
		results = append(results, s)
	}

	return results, rows.Err()
}

type TraceInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Service   string    `json:"service"`
	Duration  int       `json:"duration"` // Placeholder, will need actual trace data
	Spans     int       `json:"spans"`    // Placeholder
	Timestamp time.Time `json:"timestamp"`
	HasError  bool      `json:"has_error"`
}

func (s *Store) GetTraces(limit int) ([]TraceInfo, error) {
	// For now, group events by trace_id to simulate traces
	query := `
		SELECT
			trace_id,
			GROUP_CONCAT(name, ', ') as names,
			service,
			COUNT(*) as span_count,
			MIN(timestamp) as start_time,
			MAX(timestamp) as end_time,
			CASE WHEN SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) > 0 THEN 1 ELSE 0 END as has_error
		FROM events
		GROUP BY trace_id, service
		ORDER BY start_time DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}
	defer rows.Close()

	var results []TraceInfo
	for rows.Next() {
		var t TraceInfo
		var startTimeStr, endTimeStr string
		var hasErrorInt int
		var names string

		err := rows.Scan(&t.ID, &names, &t.Service, &t.Spans, &startTimeStr, &endTimeStr, &hasErrorInt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace info: %w", err)
		}

		startTime, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}

		endTime, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}

		t.Name = names // For simplicity, use concatenated names
		t.Timestamp = startTime
		t.HasError = hasErrorInt == 1
		t.Duration = int(endTime.Sub(startTime).Milliseconds())

		results = append(results, t)
	}

	return results, rows.Err()
}

func (s *Store) SearchEvents(query string, limit int) ([]model.Event, error) {
	// Search in name, service, and trace_id fields
	searchQuery := `
		SELECT id, timestamp, level, service, name, trace_id, data
		FROM events
		WHERE name LIKE ? OR service LIKE ? OR trace_id LIKE ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	likeQuery := "%" + query + "%"
	rows, err := s.db.Query(searchQuery, likeQuery, likeQuery, likeQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		var dataStr sql.NullString
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Service, &e.Name, &e.TraceID, &dataStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if dataStr.Valid && dataStr.String != "" {
			err = json.Unmarshal([]byte(dataStr.String), &e.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
			}
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (s *Store) GetEventsByTraceID(traceID string) ([]model.Event, error) {
	query := `
		SELECT id, timestamp, level, service, name, trace_id, data
		FROM events
		WHERE trace_id = ?
		ORDER BY timestamp ASC
	`

	rows, err := s.db.Query(query, traceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by trace ID: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		var dataStr sql.NullString
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Service, &e.Name, &e.TraceID, &dataStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if dataStr.Valid && dataStr.String != "" {
			err = json.Unmarshal([]byte(dataStr.String), &e.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
			}
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (s *Store) Close() error {
	return s.db.Close()
}
