package main

import (
	"net/http"

	// Import the core packages

	api_v1 "wispy-core/internal/api/v1"
	"wispy-core/internal/core"
	"wispy-core/internal/core/html"
	"wispy-core/internal/models"
	"wispy-core/pkg/auth"
	"wispy-core/pkg/common"

	"github.com/go-chi/chi/v5"
)

// NOTE: Page route registration is now handled by the RouteWrapper in core/route_wrapper.go

// Start initializes and starts the HTTP server
func Start(host, port, env, sitesPath, staticPath string, router *chi.Mux) {
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
	router.Handle("/public", core.StaticAndSitePublicHandler(sitesPath, staticPath))
	router.Handle("/static", core.StaticAndSitePublicHandler(sitesPath, staticPath))

	for _, instance := range siteInstances {
		instance.Mu.Lock()
		defer instance.Mu.Unlock()

		auth.CreateAuthTables(instance)

		// Use a subrouter for each site
		siteRouter := chi.NewRouter()

		// setup site proxies
		for route, target := range instance.RouteProxies {
			siteRouter.Mount(route, core.NewReverseProxyHandler(target, route))
		}

		// Set site pages
		for _, page := range instance.Pages {
			siteRouter.Get(page.Slug, func(w http.ResponseWriter, r *http.Request) {
				common.Debug("Rendering page: %s for site: %s", page.Slug, instance.Name)

				// Validate authentication and roles
				enrichedRequest, proceed := auth.ValidatePageAuthAndRoles(w, r, page, instance)
				if !proceed {
					return // Authentication failed or insufficient permissions
				}
				r = enrichedRequest

				// Render the page using the site instance
				data := map[string]interface{}{}

				// Add user to data if authenticated
				var authUser *auth.User
				authUser, _ = r.Context().Value(auth.UserContextKey).(*auth.User)

				common.Debug("- route: %s", page.Slug)

				var user = &models.UserContext{
					Email:       authUser.Email,
					FirstName:   authUser.FirstName,
					LastName:    authUser.LastName,
					DisplayName: authUser.DisplayName,
					Avatar:      authUser.Avatar,
					IsActive:    authUser.IsActive,
					Roles:       authUser.Roles,
					CreatedAt:   authUser.CreatedAt,
					UpdatedAt:   authUser.UpdatedAt,
				}

				html.RenderPageWithLayout(w, r, page, instance, user, data)
			})
		}

		siteRouter.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			common.Debug("Not found for site: %s, path: %s", instance.Name, r.URL.Path)
			// Handle 404 for the site
			common.Redirect404(w, r, "")
		}))

		instance.Router = siteRouter
	}

	// Set the NotFoundHandler for the siteRouter
	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		instance := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		instance.Router.ServeHTTP(w, r)
	}))

	// Start the HTTP server
	addr := host + ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	common.Info("Server starting on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		common.Fatal("Server failed to start: %v", err)
	}
}
