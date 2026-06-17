package model

import "time"

type Event struct {
	ID          string         `json:"id"`
	Timestamp   time.Time      `json:"timestamp"`
	Level       string         `json:"level"`
	Service     string         `json:"service"`
	Name        string         `json:"name"`
	TraceID     string         `json:"trace_id"`
	SpanID      string         `json:"span_id"`
	Environment string         `json:"environment"`
	DurationMs  float64        `json:"duration_ms"`
	Data        map[string]any `json:"data,omitempty"`
}
