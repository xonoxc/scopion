package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/xonoxc/scopion/internal/api/httpx"
	"github.com/xonoxc/scopion/internal/app/appcontext"
	"github.com/xonoxc/scopion/internal/store"
	"github.com/xonoxc/scopion/orchestrator"

	migrateable "github.com/xonoxc/scopion/internal/store/migratable"
)

type SwitchDBRequest struct {
	Dialect string `json:"dialect"`
	DSN     string `json:"dsn,omitempty"`
}

type SwitchDBResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func SwitchDBHandler(as *appcontext.AtomicAppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		reqBody := SwitchDBRequest{}
		if !httpx.DecodeJSON(w, r, &reqBody) {
			return
		}

		dialect, err := ParseDialect(reqBody.Dialect)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if dialect != migrateable.POSTGRES {
			http.Error(w, "only postgres is supported for db switching", http.StatusBadRequest)
			return
		}

		orches := orchestrator.New(as)

		err = orches.MigrateTo(store.DUAL_WRITE, reqBody.DSN)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := SwitchDBResponse{
			Status:  "ok",
			Message: "system switched to dual-write mode",
		}

		httpx.WriteJSON(w, http.StatusOK, resp)
	}
}

func ParseDialect(input string) (migrateable.DatabaseName, error) {
	if strings.TrimSpace(input) == "" {
		/*
		   default is Postgres
		*/
		return migrateable.POSTGRES, nil
	}

	d := migrateable.DatabaseName(strings.ToLower(input))
	if !d.Valid() {
		return "", errors.New("invalid database dialect")
	}

	return d, nil
}
