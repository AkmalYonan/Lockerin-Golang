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
	var targetPath string

	// 1. Check custom rewrite query param passed by vercel.json
	q := r.URL.Query()
	if p := q.Get("__path"); p != "" {
		targetPath = p
		q.Del("__path")
		r.URL.RawQuery = q.Encode()
	} else if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" && matchedPath != "/api/index" && matchedPath != "/api" && matchedPath != "/api/" {
		targetPath = matchedPath
	} else if matches := r.Header.Get("x-now-route-matches"); matches != "" {
		if vals, err := url.ParseQuery(matches); err == nil {
			if matched := vals.Get("1"); matched != "" {
				targetPath = matched
			}
		}
	} else if r.URL.Path == "/api/index" || r.URL.Path == "/api" {
		targetPath = "/"
	} else if strings.HasPrefix(r.URL.Path, "/api/index/") {
		targetPath = strings.TrimPrefix(r.URL.Path, "/api/index")
	} else {
		targetPath = r.URL.Path
	}

	if targetPath == "" {
		targetPath = "/"
	}
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}

	r.URL.Path = targetPath
	r.URL.RawPath = "" // Explicitly clear RawPath so Chi router strictly routes against targetPath
	r.RequestURI = targetPath
}
