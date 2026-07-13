package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) registerIngestionHandlers(r *mux.Router) *mux.Router {
	r.HandleFunc("/createJobs", a.createJobs).Methods("GET")

	return r
}

func (a *API) createJobs(w http.ResponseWriter, r *http.Request) {

}
