package server

import (
	"net/http"
	"time"

	// Import the core packages
	"wispy-core/internal/core"
	coreHttp "wispy-core/internal/core/http"
	"wispy-core/pkg/common"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

// NOTE: Page route registration is now handled by the RouteWrapper in core/route_wrapper.go

// Start initializes and starts the HTTP server
func Start() {
	// Get configuration from environment
	port := common.GetEnv("PORT", "8080")
	host := common.GetEnv("HOST", "localhost")
	sitesPath := common.MustGetEnv("SITES_PATH") // Required for system to function
	env := common.GetEnv("ENV", "development")

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	// Log startup information
	common.Startup("Starting Wispy Core CMS")
	common.Config("Sites directory: %s", sitesPath)
	common.Config("Environment: %s", env)
	common.Config("Host: %s, Port: %s", host, port)
	common.Config("Rate limiting: %d req/sec, %d req/min", requestsPerSecond, requestsPerMinute)

	// Load all sites and their pages from the sites directory
	common.Info("Loading sites from path: %s", common.RootSitesPath())
	sites, err := core.ImportAllSites(sitesPath)
	if err != nil {
		common.Fatal("Failed to load sites from %s: %v", common.RootSitesPath(), err)
	}

	if len(sites) == 0 {
		common.Warning("No sites were found in %s. Make sure to create at least one site directory.", common.RootSitesPath())
	} else {
		common.Success("Loaded %d sites successfully", len(sites))
	}

	// Create the main router with global middleware
	router := chi.NewRouter()

	// Apply global middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(120 * time.Second))

	// Apply rate limiting middleware
	if requestsPerSecond > 0 {
		router.Use(httprate.Limit(
			requestsPerSecond,
			1*time.Second,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			}),
		))
	}

	// Register all sites with routes and the global site registry
	siteRouterParent := chi.NewRouter()
	// Create a route wrapper around our router
	routeWrapper := coreHttp.NewRouteWrapperWithStats(siteRouterParent, true)
	// Load all sites into the route wrapper
	for _, site := range sites {
		// Register the site with the route wrapper
		routeWrapper.RegisterSite(site)
	}
	routeWrapper.RegenerateAllRoutes()
	router.Mount("/", routeWrapper)

	// Set up API routes
	router.Mount("/api", setupAPIRouter())

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
