package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveAndCleanRequest(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		headers      map[string]string
		expectedPath string
	}{
		{
			name:         "Vercel rewrite query __path=/health",
			url:          "/api/index?__path=/health",
			expectedPath: "/health",
		},
		{
			name:         "Vercel rewrite query __path=/api/v1/locations",
			url:          "/api/index?__path=/api/v1/locations&city=Jakarta",
			expectedPath: "/api/v1/locations",
		},
		{
			name:         "Vercel root rewrite __path=/",
			url:          "/api/index?__path=/",
			expectedPath: "/",
		},
		{
			name: "Vercel header x-matched-path=/swagger",
			url:  "/api/index",
			headers: map[string]string{
				"x-matched-path": "/swagger",
			},
			expectedPath: "/swagger",
		},
		{
			name:         "Fallback /api/index to /",
			url:          "/api/index",
			expectedPath: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			resolveAndCleanRequest(req)

			if req.URL.Path != tt.expectedPath {
				t.Errorf("expected path %s, got %s", tt.expectedPath, req.URL.Path)
			}
			if req.URL.RawPath != "" {
				t.Errorf("expected empty RawPath, got %s", req.URL.RawPath)
			}
			if req.URL.Query().Get("__path") != "" {
				t.Errorf("expected __path to be cleaned from query string")
			}
		})
	}
}

func TestHandlerHealthIntegration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/index?__path=/health", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}

func TestHandlerRootIntegration(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/index?__path=/", nil)
	rr := httptest.NewRecorder()

	Handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rr.Code, rr.Body.String())
	}
}
