package app

import (
	"net/http"
	"wispy-core/config"
	"wispy-core/core/site"

	"github.com/go-chi/chi/v5"
)

func RegisterAppRoutes(router chi.Router, siteManager site.SiteManager) chi.Router {
	gConfig := config.GetGlobalConfig()
	cms := NewWispyCms(siteManager)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/wispy-cms/dashboard", http.StatusFound)
	})

	// Login/logout routes (public access)
	router.Get("/login", LoginHandler(cms))
	router.Post("/login", LoginHandler(cms))
	router.Get("/logout", LogoutHandler(cms))

	// Protected routes (require authentication)
	router.Group(func(r chi.Router) {
		r.Use(gConfig.GetCoreAuthMiddleware().RequireAuth)

		r.Get("/dashboard", DashboardHandler(cms))
		r.Get("/settings", SettingsHandler(cms))
		r.Post("/settings", SettingsHandler(cms))
		r.Get("/forms", FormsHandler(cms))
		r.Get("/forms/submissions", FormSubmissionsHandler(cms))
		r.Get("/forms/submissions/{formID}", FormSubmissionByIdHandler(cms))
		r.Get("/debug", DebugHandler(cms))
	})

	return router
}
