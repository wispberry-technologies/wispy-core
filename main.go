package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"wispy-core/common"
	"wispy-core/routes"
	"wispy-core/tests"
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
		log.Printf("%sWarning: Error loading .env file: %v%s", colorYellow, err, colorReset)
	}
}

func main() {
	// Get configuration from environment
	port := common.GetEnv("PORT", "8080")
	host := common.GetEnv("HOST", "localhost")
	sitesPath := common.MustGetEnv("SITES_PATH") // Required for system to function
	env := common.GetEnv("ENV", "development")

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	if common.GetEnvBool("TEST_MODE", false) {
		tests.Run(os.Args[1:]...)
	}

	// Log startup information
	log.Printf("%s🚀 Starting Wispy Core CMS%s", colorCyan, colorReset)
	log.Printf("%s📁 Sites directory: %s%s", colorGrey, sitesPath, colorReset)
	log.Printf("%s🌍 Environment: %s%s", colorGrey, env, colorReset)
	log.Printf("%s📊 Rate limiting: %d req/sec, %d req/min%s", colorGrey, requestsPerSecond, requestsPerMinute, colorReset)

	// Initialize core components
	siteManager := common.NewSiteInstanceManager()
	renderEngine := common.NewRenderEngine(siteManager)

	// Set the render engine on the site manager
	siteManager.SetRenderEngine(renderEngine)

	// Setup routes with rate limiting configuration and auth manager
	r := routes.SetupRoutes(siteManager, renderEngine)

	// Create sites directory if it doesn't exist
	if err := os.MkdirAll(sitesPath, 0755); err != nil {
		panic(fmt.Sprintf("Error creating sites directory: %v", err))
	}

	// Setup server
	addr := fmt.Sprintf("%s:%s", host, port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(common.GetEnvInt("READ_TIMEOUT", 30)) * time.Second,
		WriteTimeout: time.Duration(common.GetEnvInt("WRITE_TIMEOUT", 30)) * time.Second,
		IdleTimeout:  time.Duration(common.GetEnvInt("IDLE_TIMEOUT", 120)) * time.Second,
	}

	// Start server
	log.Printf("%s✅ Server starting on http://%s%s", colorGreen, addr, colorReset)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%sServer failed to start: %v%s", colorRed, err, colorReset)
	}
}
