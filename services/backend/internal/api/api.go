package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/briheet/kizuna/internal/config"
	"github.com/briheet/kizuna/internal/logger"
	"github.com/gorilla/mux"
)

type API struct {
	config *config.Config
	logger *logger.Logger
}

func NewApi(
	ctx context.Context,
	config *config.Config,
	logger *logger.Logger,
) *API {

	return &API{
		config: config,
		logger: logger,
	}
}

func (a *API) Server(port int) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           a.Routes(),
		ReadHeaderTimeout: time.Duration(a.config.API.ReadHeaderTimeout) * time.Second,
		ReadTimeout:       time.Duration(a.config.API.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(a.config.API.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(a.config.API.IdleTimeout) * time.Second,
	}
}

func (a *API) Routes() *mux.Router {
	r := mux.NewRouter()

	// v1 paths
	sub := r.PathPrefix("/api/v1").Subrouter()

	sub.HandleFunc("/health", a.getHealth).Methods("GET")

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
