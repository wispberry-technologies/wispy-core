package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"golang.org/x/crypto/acme/autocert"

	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/core/apiv1"
	"wispy-core/core/site"
	"wispy-core/core/tenant/app"
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
		httpPort          = globConf.GetHttpPort()
		httpsPort         = globConf.GetHttpsPort()
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
	common.Info("» Host: %s, Port: %d", host, httpsPort)
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
	httpAddr := fmt.Sprintf("%s:%d", host, httpPort)
	httpsAddr := fmt.Sprintf("%s:%d", host, httpsPort)

	// ------------
	// # Setup routes and site management
	// ------------

	// Initialize site manager
	siteManager := site.NewSiteManager(sitesPath)

	// Auth
	authConfig := auth.DefaultConfig()
	authProvider, aProviderErr := auth.NewDefaultAuthProvider(authConfig)
	if aProviderErr != nil {
		common.Fatal("Failed to create auth provider: %v", aProviderErr)
	}
	authMiddleware := auth.NewMiddleware(authProvider, authConfig)

	// Load all sites
	sites, err := siteManager.LoadAllSites()
	if err != nil {
		common.Fatal("Failed to load sites: %v", err)
	}

	common.Info("Loaded %d sites", len(sites))

	// Setup routes for all sites
	site.ScaffoldAllTenantSites(sites)

	// Setup not found handler
	notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Site not found", http.StatusNotFound)
	})

	// Mounting API routes
	rootRouter.Route("/api", func(r chi.Router) {

		// Mount API v1 routes
		apiv1.MountApiV1(r, siteManager, authProvider, authConfig, authMiddleware)
	})

	// Mounting tenant app routes
	// This will handle routes for the Wispy CMS application
	rootRouter.Route("/wispy-cms", func(r chi.Router) {
		app.RegisterAppRoutes(r, siteManager, authProvider, authConfig, authMiddleware)
	})

	// Create host router with the site manager
	hostRouter := network.NewHostRouter(siteManager, notFoundHandler, authProvider, authConfig, authMiddleware, "localhost")

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

	// Create domain list from site domains and default domains
	siteDomains := siteManager.Domains().GetDomains()

	var httpServer *http.Server
	var httpsServer *http.Server
	var certManager *autocert.Manager
	// Create the appropriate server based on environment
	if common.IsStaging() || common.IsProduction() {
		//
		certManager, httpsServer = network.NewSSLServer(certsDir, httpsAddr, siteDomains, rootRouter)
	} else {
		// For local development, use self-signed certificates
		localServer, err := network.NewLocalSSLServer(certsDir, httpsAddr, rootRouter)
		if err != nil {
			common.Fatal("Failed to create local SSL server: %v", err)
		}
		httpsServer = localServer
		common.Info("Using local self-signed certificates for development")
	}

	//
	// Create an HTTP server that both redirects to HTTPS and handles ACME challenges
	httpServer = &http.Server{
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

	// Start HTTP server in a goroutine (only if not in local mode and port is available)
	go func() {
		common.Info("Starting HTTP server on http://%s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Warning("HTTP server failed: %v", err)
		}
	}()

	// Start HTTPS server (blocking)
	// go func() {
	common.Info("Starting HTTPS server on https://%s", httpsAddr)
	if err := httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		common.Fatal("HTTPS server failed to start: %v", err)
	}
	// }()
}
