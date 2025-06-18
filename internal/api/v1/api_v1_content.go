package api_v1

import (
	"github.com/go-chi/chi/v5"
)

func ContentRouter(r chi.Router) {
	// Content routes
	r.Route("/content", func(r chi.Router) {
		// r.Use(AuthMiddleware)
		// Content management endpoints will go here
		r.Get("/", handleAPI("ListContent"))
		r.Post("/", handleAPI("CreateContent"))
	})
}
