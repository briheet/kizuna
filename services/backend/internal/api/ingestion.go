package api

import (
	"encoding/json"
	"net/http"

	"github.com/briheet/kizuna/backend/internal/domain"
	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// This handler register registers all the jobs it will get for ingestion
func (a *API) registerIngestionHandlers(r *mux.Router) *mux.Router {
	r.HandleFunc("/createJobs", a.createJobs).Methods("POST")
	r.HandleFunc("/jobsStatus", a.jobsStatus).Methods("GET")

	return r
}

// This method will help us create jobs
// Base handler will handle all type of jobs
func (a *API) createJobs(w http.ResponseWriter, r *http.Request) {
	// First validate the request and switch to the particular type
	var req types.CreateIngestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.logger.Error("Error decoding the create jobs request body", zap.String("Err:", err.Error()))
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Validate type shyt
	if err := a.validate.Struct(req); err != nil {
		a.logger.Error("error validating jobs struct", zap.Error(err))
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Switch to particular job type
	switch req.SourceType {
	// Github jobs case
	case domain.SourceTypeGithub:
		// Get github specific cfg from the body
		var cfg types.CreateGithubJobsConfig
		if err := json.Unmarshal(req.Config, &cfg); err != nil {
			a.logger.Error("error decoding github config", zap.Error(err))
			http.Error(w, "invalid github config", http.StatusBadRequest)
			return
		}

		// Validate type shyt
		if err := a.validate.Struct(cfg); err != nil {
			a.logger.Error("error validating github config struct", zap.Error(err))
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		// Call service method to bifurcate and add jobs in db
		if err := a.ingestionService.CreateGithubJobs(r.Context(), &req, &cfg); err != nil {
			a.logger.Error("error creating github jobs", zap.Error(err))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "unsupported source_type", http.StatusBadRequest)
		return
	}

}

// This method will help us fetch jobs status
// Whether a job is in process, failed, ingested, etc
func (a *API) jobsStatus(w http.ResponseWriter, r *http.Request) {

}
