package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xonoxc/scopion/internal/api/httpx"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/model"
)

type ServerStatus struct {
	DemoEnabled bool   `json:"demo_enabled"`
	Version     string `json:"version"`
}

func StatsHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		s := as.Snapshot().Store

		stats, err := s.GetStats()
		if err != nil {
			http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, stats)
	}
}

func ErrorsByServiceHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		hoursStr := r.URL.Query().Get("hours")
		hours := 24
		if hoursStr != "" {
			if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
				hours = h
			}
		}

		s := as.Snapshot().Store

		errors, err := s.GetErrorsByService(hours)
		if err != nil {
			http.Error(w, "Failed to fetch errors by service", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, errors)
	}
}

func ServicesHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		s := as.Snapshot().Store

		services, err := s.GetServices()
		if err != nil {
			http.Error(w, "Failed to fetch services", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, services)
	}
}

func TracesHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 50
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		s := as.Snapshot().Store

		traces, err := s.GetTraces(limit)
		if err != nil {
			http.Error(w, "Failed to fetch traces", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, traces)
	}
}

func SearchHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		query := r.URL.Query().Get("q")
		if query == "" {
			json.NewEncoder(w).Encode([]model.Event{})
			return
		}

		s := as.Snapshot().Store

		events, err := s.SearchEvents(query, 50)
		if err != nil {
			http.Error(w, "Failed to search events", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, events)
	}
}

func StatusHandler(demoEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		status := ServerStatus{
			DemoEnabled: demoEnabled,
			Version:     "1.0.0",
		}

		httpx.WriteJSON(w, http.StatusOK, status)
	}
}

func ThroughputHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		hoursStr := r.URL.Query().Get("hours")
		hours := 24
		if hoursStr != "" {
			if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 {
				hours = h
			}
		}

		s := as.Snapshot().Store

		throughput, err := s.GetThroughput(hours)
		if err != nil {
			http.Error(w, "Failed to fetch throughput", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, throughput)
	}
}

func EventsHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		s := as.Snapshot().Store

		events, err := s.Recent(limit)
		if err != nil {
			http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, events)
	}
}

func TraceEventsHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		traceID := r.URL.Query().Get("trace_id")
		if traceID == "" {
			http.Error(w, "trace_id parameter is required", http.StatusBadRequest)
			return
		}

		s := as.Snapshot().Store

		events, err := s.GetEventsByTraceID(traceID)
		if err != nil {
			http.Error(w, "Failed to fetch trace events", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, events)
	}
}
