package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterAppRoutes(router chi.Router) chi.Router {
	r := router.Route("/cms-admin", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/cms-admin/dashboard", http.StatusFound)
		})
		r.Get("/login", LoginHandler)
		r.Get("/logout", LogoutHandler)
		r.Get("/dashboard", DashboardHandler)
		r.Get("/settings", SettingsHandler)
		r.Get("/forms", FormsHandler)
		r.Get("/forms/submissions", FormSubmissionsHandler)
		r.Get("/forms/submissions/{formID}", FormSubmissionByIdHandler)
	})

	return r
}
