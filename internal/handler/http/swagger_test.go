package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lockerin-backend/internal/auth"
	handlerhttp "lockerin-backend/internal/handler/http"
)

func TestSwaggerEndpoints(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 0)
	router := handlerhttp.SetupRouter(&handlerhttp.RouterConfig{
		JWTManager: jwtManager,
	})

	t.Run("GET /swagger returns Swagger UI HTML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "SwaggerUIBundle") {
			t.Fatalf("expected body to contain SwaggerUIBundle initialization")
		}
	})

	t.Run("GET /docs/openapi.yaml returns OpenAPI spec", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "openapi: 3.0.3") {
			t.Fatalf("expected OpenAPI 3.0.3 spec content")
		}
	})

	t.Run("GET /docs redirects to /swagger", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/docs", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Fatalf("expected status 301, got %d", w.Code)
		}
		if loc := w.Header().Get("Location"); loc != "/swagger" {
			t.Fatalf("expected redirect to /swagger, got %s", loc)
		}
	})
}
