package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"wispy-core/common"
	"wispy-core/core"
)

// Define colors for logging
const (
	colorCyan   = "\033[36m"
	colorGrey   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panicf("%sWarning: Error loading .env file: %v%s", colorRed, err, colorReset)
	}

	// get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}
	// Set the WISPY_CORE_ROOT environment variable to the current directory
	os.Setenv("WISPY_CORE_ROOT", currentDir)
}

// NOTE: Page route registration is now handled by the RouteWrapper in core/route_wrapper.go

func main() {
	// Get configuration from environment
	port := common.GetEnv("PORT", "8080")
	host := common.GetEnv("HOST", "localhost")
	sitesPath := common.MustGetEnv("SITES_PATH") // Required for system to function
	env := common.GetEnv("ENV", "development")

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	// Log startup information
	log.Printf("%s🚀 Starting Wispy Core CMS%s", colorCyan, colorReset)
	log.Printf("%s📁 Sites directory: %s%s", colorGrey, sitesPath, colorReset)
	log.Printf("%s🌍 Environment: %s%s", colorGrey, env, colorReset)
	log.Printf("%s🔧 Host: %s, Port: %s%s", colorGrey, host, port, colorReset)
	log.Printf("%s📊 Rate limiting: %d req/sec, %d req/min%s", colorGrey, requestsPerSecond, requestsPerMinute, colorReset)

	// Load all sites and their pages from the sites directory
	sites, err := core.LoadAllSites(sitesPath)
	if err != nil {
		log.Fatalf("Failed to load sites: %v", err)
	}

	// Set up the route wrapper
	routeWrapper := core.NewRouteWrapper()

	// Register all sites with the route wrapper
	startTime := time.Now()
	pageCount := 0
	log.Printf("[INFO] Registering page routes...")

	for _, site := range sites {
		for range site.Pages {
			pageCount++
		}
		routeWrapper.RegisterSite(site)
	}

	log.Printf("[INFO] Registered %d page routes in %s", pageCount, time.Since(startTime))

	// Start the HTTP server
	addr := host + ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: routeWrapper,
	}
	log.Printf("%s✅ Server starting on http://%s%s", colorGreen, addr, colorReset)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%sServer failed to start: %v%s", colorRed, err, colorReset)
	}
}
