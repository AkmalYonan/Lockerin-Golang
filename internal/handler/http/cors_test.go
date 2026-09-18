package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"lockerin-backend/internal/app"
)

func TestCORSPreflight_LocalhostLaravel(t *testing.T) {
	appInstance, _, err := app.NewApp()
	if err != nil {
		t.Fatalf("Failed to initialize app: %v", err)
	}

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/admin/users?limit=50&offset=0", nil)
	req.Header.Set("Origin", "http://localhost:8000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type,x-requested-with,x-csrf-token")

	rec := httptest.NewRecorder()
	appInstance.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
		t.Errorf("expected status 200 or 204, got %d", rec.Code)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:8000" && allowOrigin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin to be set for http://localhost:8000, got '%s'", allowOrigin)
	}

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" {
		t.Errorf("expected Access-Control-Allow-Headers to be set, got empty")
	}
}
