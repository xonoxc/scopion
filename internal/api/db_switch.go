package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/xonoxc/scopion/internal/api/httpx"
	"github.com/xonoxc/scopion/internal/app/appcontext"
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

		_, err := ParseDialect(reqBody.Dialect)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
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
