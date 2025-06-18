package main

import (
	"net/http"

	// Import the core packages

	api_v1 "wispy-core/internal/api/v1"
	"wispy-core/internal/core"
	"wispy-core/internal/core/html"
	"wispy-core/pkg/auth"
	"wispy-core/pkg/common"

	"github.com/go-chi/chi/v5"
)

// NOTE: Page route registration is now handled by the RouteWrapper in core/route_wrapper.go

// Start initializes and starts the HTTP server
func Start(host, port, env, sitesPath string, router *chi.Mux) {
	// Load sites
	siteInstances, err := core.LoadAllSitesAsInstances(sitesPath)
	if err != nil {
		common.Fatal("Failed to load sites: %v", err)
	}

	// Set up the router
	router.Use(auth.SiteContextMiddleware(siteInstances))

	// Mount api routes
	router.Mount("/api/v1", api_v1.Router())

	// Mount static file serving
	router.Handle("/public", core.StaticFileServingWithoutContextHandler(sitesPath))

	for _, instance := range siteInstances {
		instance.Mu.Lock()
		defer instance.Mu.Unlock()

		auth.CreateAuthTables(instance)

		router.Group(func(r chi.Router) {
			for _, page := range instance.Pages {
				r.Get(page.Slug, func(w http.ResponseWriter, r *http.Request) {
					// Render the page using the site instance
					data := map[string]interface{}{
						"Site": page.SiteDetails,
					}
					html.RenderPageWithLayout(w, r, page, instance, data)
				})
			}
		})
	}

	// Start the HTTP server
	addr := host + ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	common.Success("Server starting on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		common.Fatal("Server failed to start: %v", err)
	}
}
