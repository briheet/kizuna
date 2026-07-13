package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/db"
	"github.com/briheet/kizuna/backend/internal/logger"
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
	dbClient *db.Client,
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
		ReadHeaderTimeout: time.Duration(a.config.Api.ReadHeaderTimeout) * time.Second,
		ReadTimeout:       time.Duration(a.config.Api.ReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(a.config.Api.WriteTimeout) * time.Second,
		IdleTimeout:       time.Duration(a.config.Api.IdleTimeout) * time.Second,
	}
}

func (a *API) Routes() *mux.Router {
	r := mux.NewRouter()

	// v1 paths
	sub := r.PathPrefix("/api/v1").Subrouter()

	// Register all routes from here
	a.registerHealthHandlers(sub)
	a.registerIngestionHandlers(sub)

	return r
}
