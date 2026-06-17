package model

import "time"

type ErrorGroup struct {
	ID        int       `json:"id"`
	Message   string    `json:"message"`
	Service   string    `json:"service"`
	Location  string    `json:"location"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Count     int       `json:"count"`
	Resolved  bool      `json:"resolved"`
}
