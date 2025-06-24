package server

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

// Start initializes and starts the HTTP server
func Start(config *models.ServerConfig) error {
	// Initialize the router
	router := chi.NewRouter()

	// Set up basic routes and middleware
	setupRoutes(router, config)

	// Print banner if not in quiet mode
	if !config.QuietMode {
		printServerInfo(config)
	}

	// Start the server
	serverAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	common.Info("Starting server on %s", serverAddr)

	return http.ListenAndServe(serverAddr, router)
}

// setupRoutes configures all the routes and middleware for the server
func setupRoutes(router *chi.Mux, config *models.ServerConfig) {
	// Set up middleware
	// ...

	// Set up routes
	// ...
}

// printServerInfo displays information about the server
func printServerInfo(config *models.ServerConfig) {
	common.Info("Wispy Core Server")
	common.Info("================")
	common.Info("Version: %s", "1.0.0") // Replace with actual version
	common.Info("Host: %s", config.Host)
	common.Info("Port: %d", config.Port)
	common.Info("Environment: %s", config.Env)
	common.Info("Sites Path: %s", config.SitesPath)
}

// ServerMain is the entry point for the server when run from the CLI
func ServerMain() {
	// Define command-line flags
	portFlag := flag.Int("port", 8080, "Port to listen on")
	hostFlag := flag.String("host", "localhost", "Host to bind to")
	envFlag := flag.String("env", "development", "Environment (development, production)")
	sitesPathFlag := flag.String("sites", "data/sites", "Path to sites directory")
	quietFlag := flag.Bool("quiet", false, "Suppress output")
	helpFlag := flag.Bool("help", false, "Show help information")

	// Parse command-line flags
	flag.Parse()

	// Show help if requested
	if *helpFlag {
		showServerHelp()
		return
	}

	// Create server config
	config := &models.ServerConfig{
		Port:      *portFlag,
		Host:      *hostFlag,
		Env:       *envFlag,
		SitesPath: *sitesPathFlag,
		QuietMode: *quietFlag,
	}

	// Start the server
	err := Start(config)
	if err != nil {
		common.Error("Failed to start server: %v", err)
		os.Exit(1)
	}
}

// showServerHelp displays usage information for the server
func showServerHelp() {
	fmt.Println("Wispy Core Server")
	fmt.Println("================")
	fmt.Println("Usage: wispy-cli server [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -port=<port>         Port to listen on (default: 8080)")
	fmt.Println("  -host=<host>         Host to bind to (default: localhost)")
	fmt.Println("  -env=<environment>   Environment (development, production)")
	fmt.Println("  -sites=<path>        Path to sites directory (default: data/sites)")
	fmt.Println("  -quiet               Suppress output")
	fmt.Println("  -help                Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  wispy-cli server                   Start server with default settings")
	fmt.Println("  wispy-cli server -port=3000        Start server on port 3000")
	fmt.Println("  wispy-cli server -env=production   Start server in production mode")
}
