package main

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/joho/godotenv"

	"wispy-core/pkg/common"
)

func init() {
	// Attempt to load .env file from current directory or project root
	err := godotenv.Load(".env")
	if err != nil {
		// Try looking in parent directory (if running from cmd/server)
		err = godotenv.Load("../../.env")
		if err != nil {
			common.Fatal("Error loading .env file: %v", err)
		}
	}

	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}

	// Set WISPY_CORE_ROOT based on the current directory
	// If running from project root, keep as is
	// If running from cmd/server, go up two levels
	projectRoot := currentDir
	if filepath.Base(currentDir) == "server" && filepath.Base(filepath.Dir(currentDir)) == "cmd" {
		projectRoot = filepath.Dir(filepath.Dir(currentDir))
	}

	os.Setenv("WISPY_CORE_ROOT", projectRoot)
	common.Info("WISPY_CORE_ROOT set to: %s", projectRoot)

	// Ensure SITES_PATH exists
	sitesPath := os.Getenv("SITES_PATH")
	if sitesPath == "" {
		sitesPath = "data/sites" // Default to data/sites directory in project root
		os.Setenv("SITES_PATH", sitesPath)
	}

	// Log the current environment
	common.Info("WISPY_CORE_ROOT: %s", os.Getenv("WISPY_CORE_ROOT"))
	common.Info("SITES_PATH: %s", os.Getenv("SITES_PATH"))

	// Calculate the full sites path
	var fullSitesPath string
	if filepath.IsAbs(sitesPath) {
		fullSitesPath = sitesPath
		common.Info("Using absolute SITES_PATH: %s", fullSitesPath)
	} else {
		fullSitesPath = filepath.Join(projectRoot, sitesPath)
		common.Info("Using relative SITES_PATH: %s (full path: %s)", sitesPath, fullSitesPath)
	}

	// Create sites directory if it doesn't exist
	if _, err := os.Stat(fullSitesPath); os.IsNotExist(err) {
		common.Warning("Sites directory doesn't exist, creating: %s", fullSitesPath)
		if err := os.MkdirAll(fullSitesPath, 0755); err != nil {
			common.Fatal("Failed to create sites directory: %v", err)
		}
	}
}

func main() {
	// Get configuration from environment
	port := common.GetEnv("PORT", "8080")
	host := common.GetEnv("HOST", "localhost")
	env := common.GetEnv("ENV", "development")
	sitesPath := common.MustGetEnv("SITES_PATH") // Required for system to function

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	// Log startup information
	common.Startup("Starting Wispy Core CMS")
	common.Info("Sites directory: %s", sitesPath)
	common.Info("Environment: %s", env)
	common.Info("Host: %s, Port: %s", host, port)
	common.Info("Rate limiting: %d req/sec, %d req/min", requestsPerSecond, requestsPerMinute)

	// Create the main router with global middleware
	rootRouter := chi.NewRouter()

	// Apply global middleware
	rootRouter.Use(middleware.RequestID)
	rootRouter.Use(middleware.RealIP)
	rootRouter.Use(middleware.Logger)
	rootRouter.Use(middleware.Recoverer)
	
	// Start the server
	Start(host, port, env, sitesPath, rootRouter)
	rootRouter.Use(middleware.Timeout(120 * time.Second))

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

	// Start the server
	Start(host, port, env, sitesPath, rootRouter)
}
