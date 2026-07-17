package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/briheet/kizuna/backend/internal/config"
)

func TestRoutesHandleCORSPreflight(t *testing.T) {
	app := &API{
		config: &config.Config{
			Api: config.APIConfig{CORSAllowedOrigin: "http://localhost:4321"},
		},
	}

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/search", nil)
	rec := httptest.NewRecorder()
	app.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:4321" {
		t.Fatalf("expected configured CORS origin, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected allowed methods: %q", got)
	}
}
