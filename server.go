package main

import (
	"net/http"

	"wispy-core/common"
	"wispy-core/core"
)

// NOTE: Page route registration is now handled by the RouteWrapper in core/route_wrapper.go

func StartServer() {
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
	sites, err := core.LoadAllSites(sitesPath)
	if err != nil {
		common.Fatal("Failed to load sites: %v", err)
	}

	// Set up the route wrapper
	routeWrapper := core.NewRouteWrapper()

	// Register all sites with the route wrapper
	timer := common.StartTimer("Route registration")
	pageCount := 0
	common.Info("Registering page routes...")

	for _, site := range sites {
		for range site.Pages {
			pageCount++
		}
		routeWrapper.RegisterSite(site)
	}

	timer.EndWithMessage("Registered %d page routes", pageCount)

	// Start the HTTP server
	addr := host + ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: routeWrapper,
	}
	common.Success("Server starting on http://%s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		common.Fatal("Server failed to start: %v", err)
	}
}
