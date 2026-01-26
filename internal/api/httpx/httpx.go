package httpx

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) bool {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return false
	}

	return true
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// We intentionally ignore Encode errors here:
	// if writing fails, the connection is already broken.
	_ = json.NewEncoder(w).Encode(v)
}

func RequireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
