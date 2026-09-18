package handler

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"lockerin-backend/internal/app"
)

var (
	appInstance *app.App
	initErr     error
	once        sync.Once
)

// Handler is the Vercel serverless entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		appInstance, _, initErr = app.NewApp()
	})

	if initErr != nil {
		http.Error(w, "Lockerin API Initialization Error: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	resolveAndCleanRequest(r)

	appInstance.Router.ServeHTTP(w, r)
}

func resolveAndCleanRequest(r *http.Request) {
	// 1. Check custom rewrite query param passed by vercel.json
	q := r.URL.Query()
	if targetPath := q.Get("__path"); targetPath != "" {
		if !strings.HasPrefix(targetPath, "/") {
			targetPath = "/" + targetPath
		}
		r.URL.Path = targetPath
		q.Del("__path")
		r.URL.RawQuery = q.Encode()
		return
	}

	// 2. Check x-matched-path header (from Vercel)
	if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" && matchedPath != "/api/index" && matchedPath != "/api" && matchedPath != "/api/" {
		r.URL.Path = matchedPath
		return
	}

	// 3. Check x-now-route-matches (format: 1=health or 1=%2Fhealth)
	if matches := r.Header.Get("x-now-route-matches"); matches != "" {
		if vals, err := url.ParseQuery(matches); err == nil {
			if matched := vals.Get("1"); matched != "" {
				if !strings.HasPrefix(matched, "/") {
					matched = "/" + matched
				}
				r.URL.Path = matched
				return
			}
		}
	}

	// 4. Fallback if request reached handler as /api/index or /api
	if r.URL.Path == "/api/index" || r.URL.Path == "/api" {
		r.URL.Path = "/"
	} else if strings.HasPrefix(r.URL.Path, "/api/index/") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/index")
	}
}
