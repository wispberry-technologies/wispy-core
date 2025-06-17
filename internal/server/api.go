package server

import (
	"log"
	"net/http"

	authHandler "wispy-core/internal/api/v1/handlers/auth-handler"

	"github.com/go-chi/chi"
)

// setupAPIRouter creates and configures the API router
func setupAPIRouter() http.Handler {
	r := chi.NewRouter()

	log.Print("Setting up API router \n")

	// API versioning
	return r.Route("/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Get("/me", authHandler.WithAuth(authHandler.Me))
			r.Post("/refresh", authHandler.WithAuth(authHandler.Refresh))
			r.Post("/change-password", authHandler.WithAuth(authHandler.ChangePassword))
		})

		// OAuth routes
		r.Get("/oauth/{provider}", authHandler.OAuthRedirect)
		r.Get("/oauth/{provider}/callback", authHandler.OAuthCallback)

		// User routes
		r.Route("/users", func(r chi.Router) {
			// r.Use(AuthMiddleware)
			// User management endpoints will go here
			r.Get("/", handleAPI("ListUsers"))
			r.Post("/", handleAPI("CreateUser"))
			r.Get("/{id}", handleAPI("GetUser"))
			r.Put("/{id}", handleAPI("UpdateUser"))
			r.Delete("/{id}", handleAPI("DeleteUser"))
		})

		// Content routes
		r.Route("/content", func(r chi.Router) {
			// r.Use(AuthMiddleware)
			// Content management endpoints will go here
			r.Get("/", handleAPI("ListContent"))
			r.Post("/", handleAPI("CreateContent"))
		})
	})
}

// handleAPI is a temporary handler for API endpoints that aren't fully implemented yet
func handleAPI(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Endpoint ` + name + ` not yet fully implemented"}`))
	}
}
