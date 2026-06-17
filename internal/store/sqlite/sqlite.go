package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xonoxc/scopion/internal/model"
)

type SqliteStore struct {
	db *sql.DB
}

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

func (s *SqliteStore) Migrate() error {
	migrations := []struct {
		id    string
		sql   string
		fatal bool
	}{
		{
			id:    "001_create_events",
			fatal: true,
			sql: `CREATE TABLE IF NOT EXISTS events (
				id TEXT PRIMARY KEY,
				timestamp DATETIME NOT NULL,
				level TEXT NOT NULL,
				service TEXT NOT NULL,
				name TEXT NOT NULL,
				trace_id TEXT NOT NULL DEFAULT '',
				data TEXT DEFAULT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp DESC);
			CREATE INDEX IF NOT EXISTS idx_events_service ON events(service);
			CREATE INDEX IF NOT EXISTS idx_events_level ON events(level);
			CREATE INDEX IF NOT EXISTS idx_events_trace_id ON events(trace_id);`,
		},
		{
			id:    "002_create_spans",
			fatal: true,
			sql: `CREATE TABLE IF NOT EXISTS spans (
				trace_id TEXT NOT NULL,
				span_id TEXT NOT NULL,
				parent_span_id TEXT,
				name TEXT NOT NULL,
				service TEXT NOT NULL,
				start_time DATETIME NOT NULL,
				end_time DATETIME NOT NULL,
				status TEXT NOT NULL,
				PRIMARY KEY (trace_id, span_id)
			);`,
		},
		{
			id: "003_expand_events",
			sql: `ALTER TABLE events ADD COLUMN span_id TEXT NOT NULL DEFAULT '';
			ALTER TABLE events ADD COLUMN environment TEXT NOT NULL DEFAULT 'production';
			ALTER TABLE events ADD COLUMN duration_ms REAL NOT NULL DEFAULT 0;`,
			fatal: false,
		},
		{
			id:    "004_create_error_groups",
			fatal: true,
			sql: `CREATE TABLE IF NOT EXISTS error_groups (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				message TEXT NOT NULL,
				service TEXT NOT NULL,
				location TEXT NOT NULL DEFAULT '',
				first_seen DATETIME NOT NULL,
				last_seen DATETIME NOT NULL,
				count INTEGER NOT NULL DEFAULT 1,
				resolved INTEGER NOT NULL DEFAULT 0,
				UNIQUE(message, service, location)
			);
			CREATE INDEX IF NOT EXISTS idx_error_groups_last_seen ON error_groups(last_seen DESC);
			CREATE INDEX IF NOT EXISTS idx_error_groups_service ON error_groups(service);`,
		},
	}

	for _, m := range migrations {
		_, err := s.db.Exec(m.sql)
		if err != nil {
			errStr := err.Error()
			if m.fatal || !(strings.Contains(errStr, "duplicate column") || strings.Contains(errStr, "already exists")) {
				return fmt.Errorf("migration %s failed: %w", m.id, err)
			}
		}
	}

	return nil
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

	if e.Environment == "" {
		e.Environment = "production"
	}

	_, err = s.db.Exec(
		"INSERT INTO events (id, timestamp, level, service, name, trace_id, span_id, environment, duration_ms, data) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		e.ID, e.Timestamp, e.Level, e.Service, e.Name, e.TraceID, e.SpanID, e.Environment, e.DurationMs, string(dataJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}
	return nil
}

func (s *SqliteStore) Recent(n int) ([]model.Event, error) {
	rows, err := s.db.Query(
		"SELECT id, timestamp, level, service, name, trace_id, span_id, environment, duration_ms, data FROM events ORDER BY timestamp DESC LIMIT ?",
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
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Service, &e.Name, &e.TraceID, &e.SpanID, &e.Environment, &e.DurationMs, &dataStr)
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

	return events, nil
}

func (s *SqliteStore) GetStats() (*model.Stats, error) {
	return s.statsSince("")
}

func (s *SqliteStore) GetStatsByHours(hours int) (*model.Stats, error) {
	if hours <= 0 {
		return s.GetStats()
	}
	return s.statsSince(fmt.Sprintf(" WHERE timestamp >= datetime('now', '-%d hours')", hours))
}

func (s *SqliteStore) statsSince(where string) (*model.Stats, error) {
	var totalEvents, errorEvents, activeServices int
	err := s.db.QueryRow(fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END), 0) as errors,
			COUNT(DISTINCT service) as services
		FROM events%s
	`, where)).Scan(&totalEvents, &errorEvents, &activeServices)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := &model.Stats{
		TotalEvents:    totalEvents,
		ActiveServices: activeServices,
		ErrorEvents:    errorEvents,
	}

	if totalEvents > 0 {
		stats.ErrorRate = float64(errorEvents) / float64(totalEvents) * 100
	}

	stats.P50Latency = s.percentileSince(0.50, where)
	stats.P95Latency = s.percentileSince(0.95, where)
	stats.P99Latency = s.percentileSince(0.99, where)

	return stats, nil
}

func (s *SqliteStore) percentile(p float64) float64 {
	return s.percentileSince(p, "")
}

func (s *SqliteStore) percentileSince(p float64, where string) float64 {
	query := fmt.Sprintf("SELECT COUNT(*) FROM events WHERE duration_ms > 0%s", where)
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil || count == 0 {
		return 0
	}

	offset := int(float64(count) * p)
	if offset >= count {
		offset = count - 1
	}

	valQuery := fmt.Sprintf("SELECT duration_ms FROM events WHERE duration_ms > 0%s ORDER BY duration_ms LIMIT 1 OFFSET ?", where)
	var val float64
	err = s.db.QueryRow(valQuery, offset).Scan(&val)
	if err != nil {
		return 0
	}
	return val
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

func (s *SqliteStore) AppendSpan(span model.Span) error {
	spanID := uuid.NewString()
	_, err := s.db.Exec(
		"INSERT INTO spans (trace_id, span_id, parent_span_id, name, service, start_time, end_time, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		span.TraceID, spanID, span.ParentSpanID, span.Name, span.Service, span.StartTime, span.EndTime, span.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to insert span: %w", err)
	}
	return nil
}

func (s *SqliteStore) GetTraceSpans(traceID string) ([]model.Span, error) {
	rows, err := s.db.Query(
		"SELECT trace_id, span_id, parent_span_id, name, service, start_time, end_time, status FROM spans WHERE trace_id = ? ORDER BY start_time ASC",
		traceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query spans: %w", err)
	}
	defer rows.Close()

	var spans []model.Span
	for rows.Next() {
		var sp model.Span
		var parentSpanID sql.NullString
		err := rows.Scan(&sp.TraceID, &sp.SpanID, &parentSpanID, &sp.Name, &sp.Service, &sp.StartTime, &sp.EndTime, &sp.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan span: %w", err)
		}
		if parentSpanID.Valid {
			ps := parentSpanID.String
			sp.ParentSpanID = &ps
		}
		spans = append(spans, sp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return spans, nil
}

func (s *SqliteStore) GetTraces(limit int) ([]model.TraceInfo, error) {
	query := `
		SELECT
			trace_id,
			MIN(start_time) as start_time,
			MAX(end_time) as end_time,
			COUNT(*) as span_count,
			MAX(CASE WHEN status = 'ERROR' THEN 1 ELSE 0 END) as has_error
		FROM spans
		GROUP BY trace_id
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

		err := rows.Scan(&t.ID, &startTimeStr, &endTimeStr, &t.Spans, &hasErrorInt)
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

		t.Timestamp = startTime
		t.HasError = hasErrorInt == 1
		t.Duration = int(endTime.Sub(startTime).Milliseconds())

		results = append(results, t)
	}

	return results, rows.Err()
}

func (s *SqliteStore) SearchEvents(query string, limit int) ([]model.Event, error) {
	searchQuery := `
		SELECT id, timestamp, level, service, name, trace_id, span_id, environment, duration_ms, data
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
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Service, &e.Name, &e.TraceID, &e.SpanID, &e.Environment, &e.DurationMs, &dataStr)
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
		SELECT id, timestamp, level, service, name, trace_id, span_id, environment, duration_ms, data
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
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Service, &e.Name, &e.TraceID, &e.SpanID, &e.Environment, &e.DurationMs, &dataStr)
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

func (s *SqliteStore) UpsertErrorGroup(message, service, location string) error {
	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO error_groups (message, service, location, first_seen, last_seen, count)
		VALUES (?, ?, ?, ?, ?, 1)
		ON CONFLICT(message, service, location) DO UPDATE SET
			last_seen = excluded.last_seen,
			count = count + 1
	`, message, service, location, now, now)
	return err
}

func (s *SqliteStore) GetErrorGroups(hours int) ([]model.ErrorGroup, error) {
	rows, err := s.db.Query(`
		SELECT id, message, service, location, first_seen, last_seen, count, resolved
		FROM error_groups
		WHERE last_seen >= datetime('now', '-' || ? || ' hours')
		ORDER BY last_seen DESC
	`, hours)
	if err != nil {
		return nil, fmt.Errorf("failed to query error groups: %w", err)
	}
	defer rows.Close()

	var groups []model.ErrorGroup
	for rows.Next() {
		var g model.ErrorGroup
		var resolved int
		err := rows.Scan(&g.ID, &g.Message, &g.Service, &g.Location, &g.FirstSeen, &g.LastSeen, &g.Count, &resolved)
		if err != nil {
			return nil, fmt.Errorf("failed to scan error group: %w", err)
		}
		g.Resolved = resolved == 1
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (s *SqliteStore) DeleteEventsOlderThan(age time.Duration) (int64, error) {
	result, err := s.db.Exec(
		"DELETE FROM events WHERE timestamp < datetime('now', ?)",
		fmt.Sprintf("-%d seconds", int(age.Seconds())),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old events: %w", err)
	}
	n, _ := result.RowsAffected()
	return n, nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}
