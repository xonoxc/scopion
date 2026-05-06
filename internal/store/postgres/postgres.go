package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/google/uuid"
	"github.com/xonoxc/scopion/internal/model"
	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

type PostgresStore struct {
	Db *sql.DB
}

/**
* implementation for
*  migratable interface
**/
func (s *PostgresStore) DB() *sql.DB {
	return s.Db
}

func (s *PostgresStore) Dialect() migrateable.DatabaseName {
	return migrateable.POSTGRES
}

/*
* Implementation of Storage interface
**/
func New(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{Db: db}, nil
}

func NewWithDB(db *sql.DB) *PostgresStore {
	return &PostgresStore{Db: db}
}

func (p *PostgresStore) Append(e model.Event) error {
	var data any

	if e.Data == nil {
		jsonData, err := json.Marshal(e.Data)
		if err != nil {
			return fmt.Errorf("marshal event data: %w", err)
		}
		data = jsonData
	} else {
		data = nil
	}

	_, err := p.Db.Exec(
		`
		INSERT INTO events
		(id , timestamp , level , service , name, trace_id , data)
		VALUES ($1 , $2 , $3 , $4, $5 , $6 ,  $7)
		`,
		e.ID, e.Timestamp, e.Level, e.Service, e.Name, e.TraceID, data,
	)
	if err != nil {
		return fmt.Errorf("insert event %w:", err)
	}

	return nil
}

func (p *PostgresStore) GetStats() (*model.Stats, error) {
	var stats model.Stats

	err := p.Db.QueryRow(
		`
		SELECT
			COUNT(*) AS total_events,
			COUNT(*) FILTER (WHERE level = 'error') AS error_events,
			COUNT(DISTINCT service) AS active_services
		FROM events
		`,
	).Scan(&stats.TotalEvents, &stats.ErrorRate, &stats.ActiveServices)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	if stats.TotalEvents > 0 {
		stats.ErrorRate = stats.ErrorRate / float64(stats.TotalEvents) * 100
	}

	return &stats, nil
}

func (p *PostgresStore) GetServices() ([]model.ServiceInfo, error) {
	rows, err := p.Db.Query(
		`
		SELECT
			service,
			COUNT(*) FILTER (WHERE level = 'error') AS error_count,
			MAX(timestamp) AS last_activity,
			COUNT(*) AS event_count
		FROM events
		GROUP BY service
		ORDER BY last_activity DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var results []model.ServiceInfo
	for rows.Next() {
		var s model.ServiceInfo
		if err := rows.Scan(&s.Name, &s.ErrorCount, &s.LastActivity, &s.EventCount); err != nil {
			return nil, err
		}
		results = append(results, s)
	}

	return results, rows.Err()
}

func (p *PostgresStore) AppendSpan(span model.Span) error {
	spanID := uuid.NewString()
	_, err := p.Db.Exec(
		`
		INSERT INTO spans (trace_id, span_id, parent_span_id, name, service, start_time, end_time, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		span.TraceID, spanID, span.ParentSpanID, span.Name, span.Service, span.StartTime, span.EndTime, span.Status,
	)
	if err != nil {
		return fmt.Errorf("insert span: %w", err)
	}
	return nil
}

func (p *PostgresStore) GetTraceSpans(traceID string) ([]model.Span, error) {
	rows, err := p.Db.Query(
		`
		SELECT trace_id, span_id, parent_span_id, name, service, start_time, end_time, status
		FROM spans
		WHERE trace_id = $1
		ORDER BY start_time ASC
		`,
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
		if err := rows.Scan(&sp.TraceID, &sp.SpanID, &parentSpanID, &sp.Name, &sp.Service, &sp.StartTime, &sp.EndTime, &sp.Status); err != nil {
			return nil, fmt.Errorf("failed to scan span: %w", err)
		}
		if parentSpanID.Valid {
			ps := parentSpanID.String
			sp.ParentSpanID = &ps
		}
		spans = append(spans, sp)
	}

	return spans, rows.Err()
}

func (p *PostgresStore) GetTraces(limit int) ([]model.TraceInfo, error) {
	query := `
		SELECT
			trace_id,
			MIN(start_time) AS start_time,
			MAX(end_time) AS end_time,
			COUNT(*) AS span_count,
			BOOL_OR(status = 'ERROR') AS has_error
		FROM spans
		GROUP BY trace_id
		ORDER BY start_time DESC
		LIMIT $1
	`

	rows, err := p.Db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query traces: %w", err)
	}
	defer rows.Close()

	var results []model.TraceInfo

	for rows.Next() {
		var t model.TraceInfo
		var startTime, endTime time.Time

		err := rows.Scan(&t.ID, &startTime, &endTime, &t.Spans, &t.HasError)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trace info: %w", err)
		}

		t.Timestamp = startTime
		t.Duration = int(endTime.Sub(startTime).Milliseconds())

		results = append(results, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *PostgresStore) Recent(n int) ([]model.Event, error) {
	rows, err := p.Db.Query(
		`
		SELECT id, timestamp, level, service, name, trace_id, data
		FROM events
		ORDER BY timestamp DESC
		LIMIT $1
		`,
		n,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []model.Event

	for rows.Next() {
		var e model.Event
		var dataBytes []byte

		err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Level,
			&e.Service,
			&e.Name,
			&e.TraceID,
			&dataBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if dataBytes != nil {
			if err := json.Unmarshal(dataBytes, &e.Data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
			}
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (p *PostgresStore) GetErrorsByService(hours int) ([]model.ErrorByService, error) {
	rows, err := p.Db.Query(
		`
		SELECT service, COUNT(*) AS count
		FROM events
		WHERE level = 'error'
		  AND timestamp >= NOW() - INTERVAL '1 hour' * $1
		GROUP BY service
		ORDER BY count DESC
		`,
		hours,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query errors by service: %w", err)
	}
	defer rows.Close()

	var results []model.ErrorByService
	for rows.Next() {
		var e model.ErrorByService
		if err := rows.Scan(&e.Service, &e.Count); err != nil {
			return nil, err
		}
		results = append(results, e)
	}

	return results, rows.Err()
}

func (p *PostgresStore) SearchEvents(query string, limit int) ([]model.Event, error) {
	like := "%" + query + "%"

	rows, err := p.Db.Query(
		`
		SELECT id, timestamp, level, service, name, trace_id, data
		FROM events
		WHERE name ILIKE $1
		   OR service ILIKE $1
		   OR trace_id ILIKE $1
		ORDER BY timestamp DESC
		LIMIT $2
		`,
		like,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		var dataBytes []byte

		if err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Level,
			&e.Service,
			&e.Name,
			&e.TraceID,
			&dataBytes,
		); err != nil {
			return nil, err
		}

		if dataBytes != nil {
			_ = json.Unmarshal(dataBytes, &e.Data)
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (p *PostgresStore) GetEventsByTraceID(traceID string) ([]model.Event, error) {
	rows, err := p.Db.Query(
		`
		SELECT id, timestamp, level, service, name, trace_id, data
		FROM events
		WHERE trace_id = $1
		ORDER BY timestamp ASC
		`,
		traceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		var dataBytes []byte

		if err := rows.Scan(
			&e.ID,
			&e.Timestamp,
			&e.Level,
			&e.Service,
			&e.Name,
			&e.TraceID,
			&dataBytes,
		); err != nil {
			return nil, err
		}

		if dataBytes != nil {
			_ = json.Unmarshal(dataBytes, &e.Data)
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

func (p *PostgresStore) GetThroughput(hours int) ([]model.ThroughputData, error) {
	if hours <= 0 {
		hours = 24
	}

	rows, err := p.Db.Query(
		`
		SELECT
			date_trunc('hour', timestamp) AS time,
			COUNT(*) AS events
		FROM events
		WHERE timestamp >= NOW() - INTERVAL '1 hour' * $1
		GROUP BY time
		ORDER BY time ASC
		`,
		hours,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.ThroughputData
	for rows.Next() {
		var t model.ThroughputData
		var ts time.Time

		if err := rows.Scan(&ts, &t.Events); err != nil {
			return nil, err
		}

		t.Time = ts.Format("15:04")
		results = append(results, t)
	}

	return results, rows.Err()
}

func (s *PostgresStore) Close() error {
	return s.Db.Close()
}
