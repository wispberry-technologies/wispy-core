package common

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// ServeHTTP handles HTTP requests for sites
func (instance *SiteInstance) ServeHTTP(w http.ResponseWriter, r *http.Request, logger *Logger) {
	startTime := time.Now()

	logger.Debug("Processing request", map[string]interface{}{
		"host": r.Host,
		"path": r.URL.Path,
		"site": instance.Domain,
	})

	// Check if site is active
	if !instance.Site.IsActive {
		logger.Warn("Request to inactive site", map[string]interface{}{
			"site_domain": instance.Domain,
		})
		http.Error(w, "Site is not active", http.StatusServiceUnavailable)
		return
	}

	// Handle static assets from public directory
	if strings.HasPrefix(r.URL.Path, "/public/") {
		instance.HandleStaticAssets(w, r, logger)
		return
	}

	// Handle API requests
	if strings.HasPrefix(r.URL.Path, "/api/") {
		instance.HandleAPI(w, r, logger)
		return
	}

	// Handle page requests
	page, _, err := instance.FindRoute(r.URL.Path)
	if err != nil {
		logger.Error("Page not found", map[string]interface{}{
			"path":  r.URL.Path,
			"error": err.Error(),
		})
		http.NotFound(w, r)
		return
	}

	// Log request duration
	duration := time.Since(startTime)
	logger.Debug("Request processed", map[string]interface{}{
		"host":     r.Host,
		"path":     r.URL.Path,
		"site":     instance.Domain,
		"duration": duration.String(),
		"status":   http.StatusOK,
	})
}

// handleStaticAssets handles static asset serving
func (instance *SiteInstance) HandleStaticAssets(w http.ResponseWriter, r *http.Request, logger *Logger) {
	var assetPath string

	if strings.HasPrefix(r.URL.Path, "/public/") {
		// Public assets
		assetPath = strings.TrimPrefix(r.URL.Path, "/public/")
		logger.Debug("Serving public asset", map[string]interface{}{
			"asset_path": assetPath,
		})

		// Construct site-relative path for public assets
		sitePath := filepath.Join(MustGetEnv("SITES_PATH"), instance.Site.Domain, "public", assetPath)
		safeAssetPath, err := SecureServeFile(sitePath)
		if err != nil {
			logger.Warn("Asset access denied", map[string]interface{}{
				"path":  r.URL.Path,
				"error": err.Error(),
			})
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, safeAssetPath)
	} else {
		logger.Warn("Asset not found", map[string]interface{}{
			"path": r.URL.Path,
		})
		http.NotFound(w, r)
	}
}

// handleAPI handles API requests
func (instance *SiteInstance) HandleAPI(w http.ResponseWriter, r *http.Request, logger *Logger) {
	logger.Debug("API request received", map[string]interface{}{
		"path":   r.URL.Path,
		"method": r.Method,
	})

	// TODO: Implement API handling
	// For now, return a simple JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "API not implemented yet"}`))
}
