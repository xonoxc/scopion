package live

import (
	"encoding/json"
	"net/http"

	"github.com/xonoxc/scopion/internal/model"
)

func SSE(b *Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")

		ch := make(chan model.Event, 16)
		b.register <- ch
		defer func() { b.unregister <- ch }()

		for e := range ch {
			data, _ := json.Marshal(e)
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			w.(http.Flusher).Flush()
		}
	}
}
