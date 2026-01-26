package model

import "time"

type ServiceInfo struct {
	Name         string    `json:"name"`
	ErrorCount   int       `json:"error_count"`
	LastActivity time.Time `json:"last_activity"`
	EventCount   int       `json:"event_count"`
}
