package ingest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/pipeline"
)

type eventPayload struct {
	Level       string         `json:"level"`
	Service     string         `json:"service"`
	Name        string         `json:"name"`
	TraceID     string         `json:"trace_id,omitempty"`
	SpanID      string         `json:"span_id,omitempty"`
	Environment string         `json:"environment,omitempty"`
	DurationMs  float64        `json:"duration_ms,omitempty"`
	Data        map[string]any `json:"data,omitempty"`
}

func Handler(bp *pipeline.Batcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p eventPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		e := model.Event{
			ID:          uuid.NewString(),
			Timestamp:   time.Now(),
			Level:       p.Level,
			Service:     p.Service,
			Name:        p.Name,
			TraceID:     p.TraceID,
			SpanID:      p.SpanID,
			Environment: p.Environment,
			DurationMs:  p.DurationMs,
			Data:        p.Data,
		}

		bp.SubmitEvent(e)

		w.WriteHeader(http.StatusAccepted)
	}
}
