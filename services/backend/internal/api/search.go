package api

import (
	"encoding/json"
	"net/http"

	"github.com/briheet/kizuna/backend/internal/types"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (a *API) registerSearchHandlers(r *mux.Router) *mux.Router {
	r.HandleFunc("/search", a.search).Methods("POST")
	return r
}

func (a *API) search(w http.ResponseWriter, r *http.Request) {
	var req types.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := a.validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	resp, err := a.searchService.Search(r.Context(), req)
	if err != nil {
		a.logger.Error("search failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.logger.Error("encode search response failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
