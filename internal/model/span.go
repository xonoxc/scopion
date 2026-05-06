package model

import "time"

type Span struct {
	TraceID      string     `json:"trace_id"`
	SpanID       string     `json:"span_id"`
	ParentSpanID *string    `json:"parent_span_id,omitempty"`
	Name         string     `json:"name"`
	Service      string     `json:"service"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	Status       string     `json:"status"`
}
