package main

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"wispy-core/internal/server"
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
	// Start the server
	server.Start()
}
