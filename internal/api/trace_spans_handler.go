package api

import (
	"net/http"

	"github.com/xonoxc/scopion/internal/api/httpx"
	"github.com/xonoxc/scopion/internal/store"
)

func TraceSpansHandler(s store.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !httpx.RequireMethod(w, r, http.MethodGet) {
			return
		}

		traceID := r.URL.Query().Get("trace_id")
		if traceID == "" {
			http.Error(w, "trace_id parameter is required", http.StatusBadRequest)
			return
		}

		spans, err := s.GetTraceSpans(traceID)
		if err != nil {
			http.Error(w, "Failed to fetch trace spans", http.StatusInternalServerError)
			return
		}

		httpx.WriteJSON(w, http.StatusOK, spans)
	}
}
