package main

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strings"
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

	// Create SSL certificates directory
	certsDir := filepath.Join(projectRoot, globConf.GetCacheDir(), ".certs")
	// Create directory if it doesn't exist
	if err := common.EnsureDir(certsDir); err != nil {
		common.Warning("Failed to create certificates directory: %v", err)
	}

	// Set up default domains and get site domains
	// defaultDomains := []string{
	// 	host,        // The configured host
	// 	"*",         // Wildcard domain for unregistered domains
	// 	"localhost", // Local development
	// }

	// Create domain list from site domains and default domains
	siteDomains := siteManager.Domains().GetDomains()

	// Create the certificate manager and HTTPS server
	certManager, httpsServer := network.NewSSLServer(certsDir, httpsAddr, siteDomains, rootRouter)

	// Create an HTTP server that both redirects to HTTPS and handles ACME challenges
	httpServer := &http.Server{
		Addr: httpAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is an ACME challenge
			if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
				certManager.HTTPHandler(nil).ServeHTTP(w, r)
				return
			}

			// Otherwise redirect to HTTPS
			target := "https://" + r.Host + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

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
			common.Info("Starting HTTP server on http://%s", httpAddr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				common.Warning("HTTP server failed: %v", err)
			}
		}()
	}

	// Start HTTPS server (blocking)
	common.Info("Starting HTTPS server on https://%s", httpsAddr)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		common.Fatal("HTTPS server failed to start: %v", err)
	}
}
