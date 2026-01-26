package model

import "time"

type TraceInfo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Service   string    `json:"service"`
	Duration  int       `json:"duration"` // Placeholder, will need actual trace data
	Spans     int       `json:"spans"`    // Placeholder
	Timestamp time.Time `json:"timestamp"`
	HasError  bool      `json:"has_error"`
}
