package live

import (
	"encoding/json"
	"net/http"
	"sync"
)

func SSE(b *Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := make(chan Message, 16)
		b.Subscribe(ch)
		defer b.Unsubscribe(ch)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		var mu sync.Mutex

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
			mu.Lock()
			data, err := json.Marshal(msg)
			if err != nil {
				mu.Unlock()
				return
			}
			_, err = w.Write([]byte("data: "))
			if err != nil {
				mu.Unlock()
				return
			}
			_, err = w.Write(data)
			if err != nil {
				mu.Unlock()
				return
			}
			_, err = w.Write([]byte("\n\n"))
			if err != nil {
				mu.Unlock()
				return
			}
			mu.Unlock()
			flusher.Flush()
			}
		}
	}
}
