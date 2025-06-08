package routes

import (
	"net/http"
	"os"
	"strings"
	"time"

	"wispy-core/common"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// SetupRoutes sets up all routes for the application
func SetupRoutes(siteInstanceManager *common.SiteInstanceManager, renderEngine *common.RenderEngine) *chi.Mux {

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	r := chi.NewRouter()

	// Global middleware - add these first before any routes
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)

	// Add rate limiting middleware if enabled
	r.Use(httprate.Limit(
		requestsPerSecond, // requests
		1*time.Second,     // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))
	r.Use(httprate.LimitByIP(requestsPerMinute, time.Minute))

	// Global API routes (cross-site)
	r.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Public authentication routes
			r.Route("/auth", func(r chi.Router) {
				// Authentication endpoints
				// r.Post("/login", authHandler.Login)
				// r.Post("/register", authHandler.Register)

				// // Protected authentication endpoints
				// r.Group(func(r chi.Router) {
				// 	r.Use(authHandler.RequireAuth)
				// 	r.Post("/logout", authHandler.Logout)
				// 	r.Post("/logout-all", authHandler.LogoutAll)
				// 	r.Post("/refresh", authHandler.RefreshSession)
				// 	r.Get("/me", authHandler.GetCurrentUser)
				// 	r.Put("/profile", authHandler.UpdateProfile)
				// 	r.Post("/change-password", authHandler.ChangePassword)
				// })
			})

			// User management routes (admin only)
			r.Route("/users", func(r chi.Router) {
				// r.Use(authHandler.RequireRoles("admin"))
				// r.Get("/", authHandler.GetUsers)
				// r.Get("/{userID}", authHandler.GetUser)
				// Additional admin routes can be added here:
				// r.Put("/{userID}", authHandler.UpdateUser)
				// r.Delete("/{userID}", authHandler.DeleteUser)
				// r.Post("/{userID}/lock", authHandler.LockUser)
				// r.Post("/{userID}/unlock", authHandler.UnlockUser)
			})
		})
	})

	// Site request handler - catch all requests and route to appropriate site
	r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		// Extract domain from Host header
		domain := r.Host

		// Handle common port stripping
		if strings.Contains(domain, ":") {
			domain = strings.Split(domain, ":")[0]
		}

		// Get site instance for this domain
		siteInstance, err := siteInstanceManager.GetSiteInstance(domain)
		if err != nil {
			// Create fallback logger for error cases
			logger, logErr := common.NewLoggerFromDomain(domain, os.Getenv("SITES_PATH"))
			if logErr != nil {
				http.Error(w, "Failed to initialize logging", http.StatusInternalServerError)
				return
			}

			logger.Error("Failed to get site instance", map[string]interface{}{
				"domain": domain,
				"error":  err.Error(),
			})
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}

		// Create logger for the site instance
		logger, err := common.NewLogger(siteInstance)
		if err != nil {
			// Fallback to domain-based logger
			logger, logErr := common.NewLoggerFromDomain(domain, os.Getenv("SITES_PATH"))
			if logErr != nil {
				http.Error(w, "Failed to initialize logging", http.StatusInternalServerError)
				return
			}
			logger.Error("Failed to create site logger", map[string]interface{}{
				"domain": domain,
				"error":  err.Error(),
			})
		}

		// Route to site instance
		siteInstance.ServeHTTP(w, r, logger)
	})

	return r
}

// // setupAdminRoutes sets up admin-specific routes
// func setupAdminRoutes(r chi.Router, siteManager *common.SiteManager, renderEngine *common.RenderEngine) {
// 	// Create handlers
// 	pageHandler := handlers.NewPageHandler(siteManager, renderEngine)

// 	// Placeholder admin dashboard
// 	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "text/html")
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte(`
// 			<!DOCTYPE html>
// 			<html>
// 			<head>
// 				<title>Wispy Admin</title>
// 				<meta charset="utf-8">
// 				<meta name="viewport" content="width=device-width, initial-scale=1">
// 			</head>
// 			<body>
// 				<h1>Wispy Admin Dashboard</h1>
// 				<p>Admin interface coming soon...</p>
// 				<h2>Available APIs:</h2>
// 				<ul>
// 					<li>GET /admin/api/v1/pages?site=localhost - List pages</li>
// 					<li>GET /admin/api/v1/pages/{slug}?site=localhost - Get specific page</li>
// 					<li>POST /admin/api/v1/pages?site=localhost - Create page</li>
// 					<li>PUT /admin/api/v1/pages/{slug}?site=localhost - Update page</li>
// 					<li>DELETE /admin/api/v1/pages/{slug}?site=localhost - Delete page</li>
// 				</ul>
// 			</body>
// 			</html>
// 		`))
// 	})

// 	// API routes for admin
// 	r.Route("/api", func(r chi.Router) {
// 		r.Route("/v1", func(r chi.Router) {
// 			// Sites API
// 			r.Route("/sites", func(r chi.Router) {
// 				r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 					w.Header().Set("Content-Type", "application/json")
// 					w.WriteHeader(http.StatusNotImplemented)
// 					w.Write([]byte(`{"error": "Sites API not implemented yet"}`))
// 				})
// 			})

// 			// Pages API
// 			r.Route("/pages", func(r chi.Router) {
// 				r.Get("/", pageHandler.ListPages)
// 				r.Post("/", pageHandler.CreatePage)
// 				r.Route("/{slug}", func(r chi.Router) {
// 					r.Get("/", pageHandler.GetPage)
// 					r.Put("/", pageHandler.UpdatePage)
// 					r.Delete("/", pageHandler.DeletePage)
// 				})
// 			})
// 		})
// 	})
// }
