package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/wispberry-technologies/wispy-core/api/handlers"
	"github.com/wispberry-technologies/wispy-core/common"
)

// SetupRoutes sets up all routes for the application
func SetupRoutes(siteManager *common.SiteManager, renderEngine *common.RenderEngine, rateLimitEnabled bool, requestsPerSecond, requestsPerMinute int) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware - add these first before any routes
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// Add rate limiting middleware if enabled
	if rateLimitEnabled {
		r.Use(httprate.Limit(
			requestsPerSecond, // requests
			1*time.Second,     // per duration
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		))
		r.Use(httprate.LimitByIP(requestsPerMinute, time.Minute))
	}

	// Create handlers
	siteHandler := handlers.NewSiteHandler(siteManager, renderEngine)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	})

	// Admin routes (for managing sites, pages, etc.)
	r.Route("/admin", func(r chi.Router) {
		// TODO: Add admin authentication middleware
		setupAdminRoutes(r, siteManager, renderEngine)
	})

	// Site routes - catch all for individual sites
	r.NotFound(siteHandler.ServeHTTP)

	return r
}

// setupAdminRoutes sets up admin-specific routes
func setupAdminRoutes(r chi.Router, siteManager *common.SiteManager, renderEngine *common.RenderEngine) {
	// Create handlers
	pageHandler := handlers.NewPageHandler(siteManager, renderEngine)

	// Placeholder admin dashboard
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Wispy Admin</title>
				<meta charset="utf-8">
				<meta name="viewport" content="width=device-width, initial-scale=1">
			</head>
			<body>
				<h1>Wispy Admin Dashboard</h1>
				<p>Admin interface coming soon...</p>
				<h2>Available APIs:</h2>
				<ul>
					<li>GET /admin/api/v1/pages?site=localhost - List pages</li>
					<li>GET /admin/api/v1/pages/{slug}?site=localhost - Get specific page</li>
					<li>POST /admin/api/v1/pages?site=localhost - Create page</li>
					<li>PUT /admin/api/v1/pages/{slug}?site=localhost - Update page</li>
					<li>DELETE /admin/api/v1/pages/{slug}?site=localhost - Delete page</li>
				</ul>
			</body>
			</html>
		`))
	})

	// API routes for admin
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Sites API
			r.Route("/sites", func(r chi.Router) {
				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotImplemented)
					w.Write([]byte(`{"error": "Sites API not implemented yet"}`))
				})
			})

			// Pages API
			r.Route("/pages", func(r chi.Router) {
				r.Get("/", pageHandler.ListPages)
				r.Post("/", pageHandler.CreatePage)
				r.Route("/{slug}", func(r chi.Router) {
					r.Get("/", pageHandler.GetPage)
					r.Put("/", pageHandler.UpdatePage)
					r.Delete("/", pageHandler.DeletePage)
				})
			})
		})
	})
}
