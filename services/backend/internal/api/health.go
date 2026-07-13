package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func (a *API) registerHealthHandlers(r *mux.Router) *mux.Router {
	r.HandleFunc("/health", a.getHealth).Methods("GET")

	return r
}

func (a *API) getHealth(w http.ResponseWriter, r *http.Request) {
	a.logger.Info("Health handler reached")

	w.Header().Set("Content-type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})

	if err != nil {
		a.logger.Error("encode health response failed")
	}
}
