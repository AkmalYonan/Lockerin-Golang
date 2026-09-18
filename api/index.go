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

	extractFromStr := func(raw string) string {
		if raw == "" {
			return ""
		}
		if u, err := url.Parse(raw); err == nil {
			if p := u.Query().Get("__path"); p != "" {
				return p
			}
			if u.Path != "" && u.Path != "/api/index" && u.Path != "/api" && u.Path != "/api/" {
				return u.Path
			}
		}
		return ""
	}

	// 1. Check direct query param on r.URL
	q := r.URL.Query()
	if p := q.Get("__path"); p != "" {
		targetPath = p
		q.Del("__path")
		r.URL.RawQuery = q.Encode()
	}

	// 2. Check x-matched-path header (e.g. /api/index?__path=/health or /health)
	if targetPath == "" {
		targetPath = extractFromStr(r.Header.Get("x-matched-path"))
	}

	// 3. Check x-now-route-matches header (e.g. 1=health or 1=%2Fhealth)
	if targetPath == "" {
		if matches := r.Header.Get("x-now-route-matches"); matches != "" {
			if vals, err := url.ParseQuery(matches); err == nil {
				if matched := vals.Get("1"); matched != "" {
					targetPath = matched
				}
			}
		}
	}

	// 4. Check additional Vercel routing headers
	if targetPath == "" {
		targetPath = extractFromStr(r.Header.Get("x-invoke-path"))
	}
	if targetPath == "" {
		targetPath = extractFromStr(r.Header.Get("x-forwarded-url"))
	}
	if targetPath == "" {
		targetPath = extractFromStr(r.Header.Get("x-real-url"))
	}

	// 5. Fallback from r.URL.Path
	if targetPath == "" {
		targetPath = r.URL.Path
	}

	// Strip query string if embedded in targetPath
	if strings.Contains(targetPath, "?") {
		parts := strings.SplitN(targetPath, "?", 2)
		targetPath = parts[0]
	}

	// Strip /api/index prefix if still present
	if targetPath == "/api/index" || targetPath == "/api" || targetPath == "/api/" {
		targetPath = "/"
	} else if strings.HasPrefix(targetPath, "/api/index/") {
		targetPath = strings.TrimPrefix(targetPath, "/api/index")
	}

	if targetPath == "" {
		targetPath = "/"
	}
	if !strings.HasPrefix(targetPath, "/") {
		targetPath = "/" + targetPath
	}

	r.URL.Path = targetPath
	r.URL.RawPath = "" // Explicitly clear RawPath so Chi router strictly uses targetPath
	r.RequestURI = targetPath
}
