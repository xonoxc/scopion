package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xonoxc/scopion/internal/model"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

type SqliteStore struct {
	db *sql.DB
}

/**
* implementation for
*  migratable interface
**/
func (s *SqliteStore) DB() *sql.DB {
	return s.db
}

func (s *SqliteStore) Dialect() migrateable.DatabaseName {
	return migrateable.SQLITE
}

/*
* Implementation of Storage interface
**/
func New(dbPath string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SqliteStore{db: db}, nil
}

func NewWithDB(db *sql.DB) *SqliteStore {
	return &SqliteStore{db: db}
}

func (s *SqliteStore) Append(e model.Event) error {
	var dataJSON []byte
	var err error
	if e.Data != nil {
		dataJSON, err = json.Marshal(e.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}
	} else {
		dataJSON = nil
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

func (s *SqliteStore) Recent(n int) ([]model.Event, error) {
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

func (s *SqliteStore) GetStats() (*model.Stats, error) {
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

	return &model.Stats{
		TotalEvents:    totalEvents,
		ErrorRate:      errorRate,
		ActiveServices: activeServices,
	}, nil
}

func (s *SqliteStore) GetErrorsByService(hours int) ([]model.ErrorByService, error) {
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

	var results []model.ErrorByService
	for rows.Next() {
		var e model.ErrorByService
		err := rows.Scan(&e.Service, &e.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan error by service: %w", err)
		}
		results = append(results, e)
	}

	return results, rows.Err()
}

func (s *SqliteStore) GetServices() ([]model.ServiceInfo, error) {
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

	var results []model.ServiceInfo
	for rows.Next() {
		var s model.ServiceInfo
		err := rows.Scan(&s.Name, &s.ErrorCount, &s.LastActivity, &s.EventCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service info: %w", err)
		}
		results = append(results, s)
	}

	return results, rows.Err()
}

func (s *SqliteStore) GetTraces(limit int) ([]model.TraceInfo, error) {
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

	var results []model.TraceInfo
	for rows.Next() {
		var t model.TraceInfo
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

func (s *SqliteStore) SearchEvents(query string, limit int) ([]model.Event, error) {
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

func (s *SqliteStore) GetEventsByTraceID(traceID string) ([]model.Event, error) {
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

func (s *SqliteStore) GetThroughput(hours int) ([]model.ThroughputData, error) {
	if hours <= 0 {
		hours = 24
	}

	// Calculate events per hour for the last N hours
	query := `
		WITH hours AS (
			SELECT strftime('%Y-%m-%d %H:00:00', datetime('now', '-' || (t.n * 1) || ' hours')) as hour_start
			FROM (SELECT 0 as n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9 UNION ALL SELECT 10 UNION ALL SELECT 11 UNION ALL SELECT 12 UNION ALL SELECT 13 UNION ALL SELECT 14 UNION ALL SELECT 15 UNION ALL SELECT 16 UNION ALL SELECT 17 UNION ALL SELECT 18 UNION ALL SELECT 19 UNION ALL SELECT 20 UNION ALL SELECT 21 UNION ALL SELECT 22 UNION ALL SELECT 23) t
			WHERE t.n < ?
		)
		SELECT
			strftime('%H:00', h.hour_start) as time,
			COUNT(e.id) as events
		FROM hours h
		LEFT JOIN events e ON e.timestamp >= h.hour_start AND e.timestamp < datetime(h.hour_start, '+1 hour')
		GROUP BY h.hour_start
		ORDER BY h.hour_start ASC
	`

	rows, err := s.db.Query(query, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query throughput: %w", err)
	}
	defer rows.Close()

	var results []model.ThroughputData
	for rows.Next() {
		var t model.ThroughputData
		err := rows.Scan(&t.Time, &t.Events)
		if err != nil {
			return nil, fmt.Errorf("failed to scan throughput data: %w", err)
		}
		results = append(results, t)
	}

	return results, rows.Err()
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}
