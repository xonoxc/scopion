package model

type Stats struct {
	TotalEvents    int     `json:"total_events"`
	TotalRequests  int     `json:"total_requests"`
	ErrorEvents    int     `json:"error_events"`
	ErrorRate      float64 `json:"error_rate"`
	ActiveServices int     `json:"active_services"`
	P50Latency     float64 `json:"p50_latency"`
	P95Latency     float64 `json:"p95_latency"`
	P99Latency     float64 `json:"p99_latency"`
}
