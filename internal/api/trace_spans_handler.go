package api

import (
	"net/http"

	"github.com/xonoxc/scopion/internal/api/httpx"
	"github.com/xonoxc/scopion/internal/app/appcontext"
)

func TraceSpansHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
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

		spans, err := s.GetTraceSpans(traceID)
		if err != nil {
			http.Error(w, "Failed to fetch trace spans", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, spans)
	}
}
