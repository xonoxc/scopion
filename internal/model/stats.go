package model

type Stats struct {
	TotalEvents    int     `json:"total_events"`
	ErrorRate      float64 `json:"error_rate"`
	ActiveServices int     `json:"active_services"`
}
