package ingest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/pipeline"
)

type spanPayload struct {
	TraceID      string     `json:"trace_id"`
	ParentSpanID *string    `json:"parent_span_id,omitempty"`
	Name         string     `json:"name"`
	Service      string     `json:"service"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Status       string     `json:"status"`
}

func SpanHandler(bp *pipeline.Batcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p spanPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		now := time.Now()
		startTime := now
		if p.StartTime != nil {
			startTime = *p.StartTime
		}
		endTime := now
		if p.EndTime != nil {
			endTime = *p.EndTime
		}

		span := model.Span{
			TraceID:      p.TraceID,
			SpanID:       uuid.NewString(),
			ParentSpanID: p.ParentSpanID,
			Name:         p.Name,
			Service:      p.Service,
			StartTime:    startTime,
			EndTime:      endTime,
			Status:       p.Status,
		}

		bp.SubmitSpan(span)

		w.WriteHeader(http.StatusAccepted)
	}
}
