package api_v1

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Router creates and configures the API router
func Router() http.Handler {
	r := chi.NewRouter()

	log.Print("Setting up API router \n")

	// API versioning
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to the API v1"}`))
	})

	// Add import and add auth router
	AuthRouter(r)
	PagesRouter(r)
	UserRouter(r)

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<h1>404 Not Found</h1><p>Are you sure you should be accessing this API?</p>`))
	})

	return r
}

// handleAPI is a temporary handler for API endpoints that aren't fully implemented yet
func handleAPI(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Endpoint ` + name + ` not yet fully implemented"}`))
	}
}
