package handler

import (
	"net/http"
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

	// Support for Vercel internal path rewrites if needed
	if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" && r.URL.Path == "/api/index" {
		r.URL.Path = matchedPath
	}

	appInstance.Router.ServeHTTP(w, r)
}
