package main

import (
	"context"
	"net/http"

	api_v1 "wispy-core/internal/api/v1"
	"wispy-core/internal/core"
	"wispy-core/pkg/auth"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

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

		// Use a subrouter for each site
		siteRouter := chi.NewRouter()

		// setup site proxies
		for route, target := range instance.RouteProxies {
			siteRouter.Mount(route, core.NewReverseProxyHandler(target, route))
		}

		// Set site pages
		for _, page := range instance.Pages {
			// Page handler function
			pageHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Create the page data
				pageData := &core.PageData{
					Template: page.FilePath,
					Data:    make(map[string]interface{}),
					Title:   page.Title,
					Path:    r.URL.Path,
				}

				// Add instance and page to context
				r = r.WithContext(context.WithValue(r.Context(), core.InstanceKey, instance))
				r = r.WithContext(context.WithValue(r.Context(), core.PageKey, pageData))

				// Debug logging
				contextData := map[string]interface{}{
					"Site":       instance,
					"Instance":   instance,
					"Page":       pageData,
					"AuthConfig": instance.AuthConfig,
				}
				common.Debug("Template context data prepared",
					"contextKeys", common.GetMapKeys(contextData),
					"template", page.FilePath)

				// Cast instance to core.Instance and render template
				if coreInstance, ok := instance.(*core.Instance); ok && coreInstance.TemplateEngine != nil {
					err := coreInstance.TemplateEngine.RenderTemplate(w, page.FilePath, contextData)
					if err != nil {
						common.Error("Template render failed: %v", err)
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
				} else {
					common.Error("Invalid instance type or template engine is nil for instance %s", instance.Domain)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			})

			siteRouter.Get(page.Slug, pageHandler)
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

	common.Success("Server starting on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		common.Fatal("Server failed to start: %v", err)
	}
}

// createInstanceContext adds instance and page data to the request context
func createInstanceContext(r *http.Request, instance *core.Instance, page *core.PageData) *http.Request {
	// Add instance to context
	ctx := context.WithValue(r.Context(), core.InstanceKey, instance)
	// Add page to context
	ctx = context.WithValue(ctx, core.PageKey, page)
	// Return new request with updated context
	return r.WithContext(ctx)
}
