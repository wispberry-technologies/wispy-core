package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/core/site"
	"wispy-core/network"
)

func main() {
	// -----------
	// # Load configuration
	// -----------

	globConf := config.LoadGlobalConfig() // Load global configuration

	// Get current working directory
	var (
		projectRoot       = globConf.GetProjectRoot()
		port              = globConf.GetPort()
		host              = globConf.GetHost()
		env               = globConf.GetEnv()
		sitesPath         = globConf.GetSitesPath()
		staticPath        = globConf.GetStaticPath()
		requestsPerSecond = globConf.GetRequestsPerSecond()
		requestsPerMinute = globConf.GetRequestsPerMinute()
	)
	// ------------
	// # Initialize router
	// # Set up middleware
	// - Request ID
	// - Real IP
	// - Recovery
	// - Timeout
	// - Rate limiting
	// - Request logging
	// ------------

	// Log startup information
	common.Info("Starting Wispy Core CMS")
	common.Info("» Project root: %s", projectRoot)
	common.Info("» Sites directory: %s", sitesPath)
	common.Info("» Static directory: %s", staticPath)
	common.Info("» Environment: %s", env)
	common.Info("» Host: %s, Port: %d", host, port)
	common.Info("» Rate limiting: %d req/sec, %d req/min", requestsPerSecond, requestsPerMinute)

	// Create the main router with global middleware
	rootRouter := chi.NewRouter()

	// Apply global middleware
	rootRouter.Use(middleware.RequestID)
	rootRouter.Use(middleware.RealIP)
	rootRouter.Use(middleware.Recoverer)
	rootRouter.Use(middleware.Timeout(120 * time.Second))
	// TODO: Add more security middleware as needed
	// Create new security middleware in security package
	// REF: https://github.com/unrolled/secure
	rootRouter.Use(common.RequestLogger()) // Log all requests

	// Apply rate limiting middleware
	if requestsPerSecond > 0 {
		rootRouter.Use(httprate.Limit(
			requestsPerSecond,
			1*time.Second,
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
			}),
		))
	}

	certsDir := filepath.Join(globConf.GetProjectRoot(), globConf.GetCacheDir(), "/.certs")
	domains := network.NewDomainList()
	// Add default domains if needed
	defaultDomains := []string{"localhost", "example.com", "www.example.com"}
	for _, domain := range defaultDomains {
		if err := domains.AddDomain(domain); err != nil {
			common.Warning("Failed to add default domain %q: %v", domain, err)
		}
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	// TODO: api support for handling adding domains dynamically
	_, server := network.NewSSLServer(certsDir, addr, domains, rootRouter)

	// ------------
	// # Setup routes and site management
	// ------------

	// Initialize site manager
	siteManager := site.NewSiteManager(sitesPath)

	// Load all sites
	sites, err := siteManager.LoadAllSites()
	if err != nil {
		common.Fatal("Failed to load sites: %v", err)
	}

	common.Info("Loaded %d sites", len(sites))

	// Setup routes for all sites
	site.ScaffoldAllSites(sites)

	// Setup not found handler
	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Site not found", http.StatusNotFound)
	})

	// Create host router with the site manager
	hostRouter := network.NewHostRouter(siteManager, notFoundHandler, "localhost")

	// Set the host router as the main handler
	rootRouter.Mount("/", hostRouter)

	// Setup static file serving
	rootRouter.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Start the HTTP server
	common.Info("Server starting on https://%s", addr)
	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		common.Fatal("Server failed to start: %v", err)
	}
}
