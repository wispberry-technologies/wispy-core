package main

import (
	"fmt"
	"net"
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

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

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

	// Use standard ports for HTTP and HTTPS
	httpAddr := fmt.Sprintf("%s:80", host)   // Standard HTTP port
	httpsAddr := fmt.Sprintf("%s:443", host) // Standard HTTPS port

	// If a custom port was specified, use it for HTTPS
	if port != 80 && port != 443 {
		httpsAddr = fmt.Sprintf("%s:%d", host, port)
	}

	// Create an HTTP server that redirects to HTTPS
	httpServer := &http.Server{
		Addr: httpAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Build the redirect URL
			target := "https://" + r.Host + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}

			// Redirect to HTTPS
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

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

	// Create SSL certificates directory
	certsDir := filepath.Join(globConf.GetProjectRoot(), globConf.GetCacheDir(), "/.certs")

	// Set up default domains and get site domains
	defaultDomains := []string{
		host,        // The configured host
		"*",         // Wildcard domain for unregistered domains
		"localhost", // Local development
	}

	// Create domain list from site domains and default domains
	siteDomains := siteManager.Domains().GetDomains()
	domains := network.CreateDomainListFromMap(siteDomains, defaultDomains)
	_, httpsServer := network.NewSSLServer(certsDir, httpsAddr, domains, rootRouter)

	// Create host router with the site manager
	hostRouter := network.NewHostRouter(siteManager, notFoundHandler, "localhost")

	// Set the host router as the main handler
	rootRouter.Mount("/", hostRouter)

	// Setup static file serving
	rootRouter.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Start both HTTP and HTTPS servers
	common.Info("HTTP server starting on http://%s", httpAddr)
	common.Info("HTTPS server starting on https://%s", httpsAddr)

	// Channel to collect errors from the HTTP server
	httpErrors := make(chan error, 1)

	// Extract port numbers from addresses for availability checking
	httpPort := 80
	httpsPort := 443
	if port != 80 && port != 443 {
		httpsPort = port
	}

	// Test if ports are available before starting servers
	httpPortAvailable := isPortAvailable(httpPort)
	httpsPortAvailable := isPortAvailable(httpsPort)

	if !httpPortAvailable {
		common.Warning("HTTP port %d is already in use. HTTP redirects will not be available.", httpPort)
	}

	if !httpsPortAvailable {
		common.Fatal("HTTPS port %d is already in use. Cannot start server.", httpsPort)
	}

	// Start HTTP server in a goroutine (only if port is available)
	if httpPortAvailable {
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				common.Warning("HTTP server failed: %v", err)
				httpErrors <- err
			}
		}()

		// Give the HTTP server a moment to bind to the port and report any errors
		select {
		case err := <-httpErrors:
			common.Warning("HTTP server couldn't start: %v", err)
			// Continue even if HTTP server fails, as HTTPS might still work
		case <-time.After(500 * time.Millisecond):
			// Continue after a short delay if no immediate errors
		}
	}

	// Start HTTPS server (blocking)
	common.Info("Starting HTTPS server on %s", httpsAddr)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		common.Fatal("HTTPS server failed to start: %v", err)
	}
}
