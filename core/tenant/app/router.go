package app

import (
	"net/http"
	"wispy-core/auth"
	"wispy-core/core/site"

	"github.com/go-chi/chi/v5"
)

func RegisterAppRoutes(router chi.Router, siteManager site.SiteManager, authProvider auth.AuthProvider, authConfig auth.Config, authMiddleware *auth.Middleware) chi.Router {

	cms := NewWispyCms()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/wispy-cms/dashboard", http.StatusFound)
	})
	router.Get("/login", LoginHandler(cms))
	router.Get("/logout", LogoutHandler(cms))
	router.Get("/dashboard", DashboardHandler(cms))
	router.Get("/settings", SettingsHandler(cms))
	router.Get("/forms", FormsHandler(cms))
	// router.Get("/forms/submit", FormSubmissionHandler(cms))
	router.Get("/forms/submissions", FormSubmissionsHandler(cms))
	router.Get("/forms/submissions/{formID}", FormSubmissionByIdHandler(cms))

	return router
}
