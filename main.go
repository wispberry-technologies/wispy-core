package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/wispberry-technologies/wispy-core/api/routes"
	"github.com/wispberry-technologies/wispy-core/common"
)

// Define colors for logging
const (
	colorCyan   = "\033[36m"
	colorGrey   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
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
	sitesPath := common.GetEnv("SITES_PATH", "./sites")
	env := common.GetEnv("ENV", "development")

	// Enable rate limiting based on config
	rateLimitEnabled := common.GetEnvBool("RATE_LIMIT_ENABLED", true)
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	// Log startup information
	log.Printf("%s🚀 Starting Wispy Core CMS%s", colorCyan, colorReset)
	log.Printf("%s📁 Sites directory: %s%s", colorGrey, sitesPath, colorReset)
	log.Printf("%s🌍 Environment: %s%s", colorGrey, env, colorReset)
	log.Printf("%s🔧 Rate limiting: %v%s", colorGrey, rateLimitEnabled, colorReset)

	// Initialize core components
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)

	// Setup routes with rate limiting configuration
	r := routes.SetupRoutes(siteManager, renderEngine, rateLimitEnabled, requestsPerSecond, requestsPerMinute)

	if rateLimitEnabled {
		log.Printf("%s📊 Rate limiting: %d req/sec, %d req/min%s", colorGrey, requestsPerSecond, requestsPerMinute, colorReset)
	}

	// Create sites directory if it doesn't exist
	if err := os.MkdirAll(sitesPath, 0755); err != nil {
		log.Fatalf("Error creating sites directory: %v", err)
	}

	// Setup server
	addr := fmt.Sprintf("%s:%s", host, port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	log.Printf("%s✅ Server starting on http://%s%s", colorGreen, addr, colorReset)
	log.Printf("%s🔗 Admin interface: http://%s/admin%s", colorGreen, addr, colorReset)
	log.Printf("%s💚 Health check: http://%s/health%s", colorGreen, addr, colorReset)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
