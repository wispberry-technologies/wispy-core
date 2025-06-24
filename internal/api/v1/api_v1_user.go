package api_v1

import (
	"github.com/go-chi/chi/v5"
)

func UserRouter(r chi.Router) {
	// User routes
	r.Route("/users", func(r chi.Router) {
		// User management endpoints will go here
		r.Get("/", handleAPI("ListUsers"))
		r.Post("/", handleAPI("CreateUser"))
		r.Get("/{id}", handleAPI("GetUser"))
		r.Put("/{id}", handleAPI("UpdateUser"))
		r.Delete("/{id}", handleAPI("DeleteUser"))
	})
}
