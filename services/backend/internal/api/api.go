package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/briheet/kizuna/backend/internal/config"
	"github.com/briheet/kizuna/backend/internal/db"
	"github.com/briheet/kizuna/backend/internal/logger"
	"github.com/briheet/kizuna/backend/internal/repository/cockroachdb"
	"github.com/briheet/kizuna/backend/internal/repository/embedder"
	"github.com/briheet/kizuna/backend/internal/services"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type API struct {
	config   *config.Config
	logger   *logger.Logger
	validate *validator.Validate

	ingestionService *services.IngestionService
	searchService    *services.SearchService
}

func NewApi(
	ctx context.Context,
	config *config.Config,
	logger *logger.Logger,
	dbClient *db.Client,
) *API {

	// Ingestion service init
	ingestionRepo := cockroachdb.NewCockroachDbIngestionRepository(dbClient)
	ingestionService := services.NewIngestionService(ingestionRepo)
	searchRepo := cockroachdb.NewCockroachDbSearchRepository(dbClient)
	embedderRepo := embedder.NewNomicRepository(config.Embedder.BaseURL)
	searchService := services.NewSearchService(searchRepo, embedderRepo)

	return &API{
		config:   config,
		logger:   logger,
		validate: validator.New(validator.WithRequiredStructEnabled()),

		ingestionService: ingestionService,
		searchService:    searchService,
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
	a.registerSearchHandlers(sub)

	return r
}
