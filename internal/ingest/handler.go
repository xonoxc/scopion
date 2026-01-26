package ingest

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store"

	"github.com/google/uuid"
)

func Handler(store store.Storage, live *live.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var e model.Event
		if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		e.ID = uuid.NewString()
		e.Timestamp = time.Now()

		store.Append(e)
		live.Publish(e)

		w.WriteHeader(http.StatusAccepted)
	}
}
