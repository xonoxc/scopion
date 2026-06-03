package ingest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/model"
)

func Handler(as *appcontext.AtomicAppState, live *live.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e model.Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		e.ID = uuid.NewString()
		e.Timestamp = time.Now()

		store := as.Snapshot().Store
		store.Append(e)
		live.Publish(e)

		w.WriteHeader(http.StatusAccepted)
	}
}
