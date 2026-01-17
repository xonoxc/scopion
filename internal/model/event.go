package model

import "time"

type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Service   string    `json:"service"`
	Name      string    `json:"name"`
	TraceID   string    `json:"trace_id"`
}
