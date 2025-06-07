package handlers

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/wispberry-technologies/wispy-core/common"
)

// SiteHandler handles requests for individual sites
type SiteHandler struct {
	siteManager   *common.SiteManager
	renderEngine  *common.RenderEngine
	loggerManager *common.LoggerManager
}

// NewSiteHandler creates a new site handler
func NewSiteHandler(siteManager *common.SiteManager, renderEngine *common.RenderEngine, loggerManager *common.LoggerManager) *SiteHandler {
	return &SiteHandler{
		siteManager:   siteManager,
		renderEngine:  renderEngine,
		loggerManager: loggerManager,
	}
}

// ServeHTTP handles HTTP requests for sites
func (sh *SiteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get site from host header
	site, err := sh.siteManager.GetSiteFromHost(r.Host)
	if err != nil {
		// Use a default logger if site is not found
		defaultLogger := sh.loggerManager.GetDefaultLogger()
		defaultLogger.Error("Failed to get site", map[string]interface{}{
			"host":  r.Host,
			"path":  r.URL.Path,
			"error": err.Error(),
		})
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Get logger for this site
	logger, err := sh.loggerManager.GetLogger(site.Domain)
	if err != nil {
		// Fallback to default logger if site logger cannot be created
		logger = sh.loggerManager.GetDefaultLogger()
		logger.Error("Failed to get logger for site", map[string]interface{}{
			"site_domain": site.Domain,
			"error":       err.Error(),
		})
	}

	logger.Debug("Processing request", map[string]interface{}{
		"host": r.Host,
		"path": r.URL.Path,
		"site": site.Domain,
	})

	// Check if site is active
	if !site.IsActive {
		logger.Warn("Request to inactive site", map[string]interface{}{
			"site_domain": site.Domain,
		})
		http.Error(w, "Site is not active", http.StatusServiceUnavailable)
		return
	}

	// Handle static assets
	if strings.HasPrefix(r.URL.Path, "/assets/") || strings.HasPrefix(r.URL.Path, "/public/") {
		sh.handleStaticAssets(w, r, site, logger)
		return
	}

	// Handle API requests
	if strings.HasPrefix(r.URL.Path, "/api/") {
		sh.handleAPI(w, r, site, logger)
		return
	}

	// Handle page requests
	sh.handlePage(w, r, site, logger)

	// Log request duration
	duration := time.Since(startTime)
	logger.Debug("Request processed", map[string]interface{}{
		"host":     r.Host,
		"path":     r.URL.Path,
		"site":     site.Domain,
		"duration": duration.String(),
		"status":   http.StatusOK,
	})
}

// handlePage handles page rendering with dynamic routing
func (sh *SiteHandler) handlePage(w http.ResponseWriter, r *http.Request, site *common.Site, logger *common.Logger) {
	// Get route manager for this site
	routeManager := common.NewRouteManager(site, logger)

	// Populate routes from site pages
	if err := sh.populateRoutes(site, routeManager, logger); err != nil {
		logger.Error("Failed to populate routes", map[string]interface{}{
			"error": err.Error(),
		})
		sh.renderError(w, r, site, logger, http.StatusInternalServerError, "Error loading routes")
		return
	}

	// Match route
	route, params, err := routeManager.FindRoute(r.URL.Path)
	if err != nil {
		logger.Debug("No route matched", map[string]interface{}{
			"path": r.URL.Path,
		})
		sh.renderError(w, r, site, logger, http.StatusNotFound, "Page not found")
		return
	}

	// Create page manager
	pageManager := common.NewPageManager(site)

	// Get page by slug from route
	page, err := pageManager.GetPage(route.PageSlug)
	if err != nil {
		logger.Error("Failed to get page", map[string]interface{}{
			"page_slug": route.PageSlug,
			"error":     err.Error(),
		})
		sh.renderError(w, r, site, logger, http.StatusNotFound, "Page not found")
		return
	}

	logger.Debug("Found page", map[string]interface{}{
		"title":    page.Meta.Title,
		"slug":     route.PageSlug,
		"url":      page.Meta.URL,
		"template": page.Meta.Template,
	})

	// Check if page is draft and user is not authenticated admin
	if page.Meta.IsDraft {
		// TODO: Implement admin authentication check
		logger.Debug("Serving draft page", map[string]interface{}{
			"page_slug": route.PageSlug,
		})
	}

	// Check if page requires authentication
	if page.Meta.RequireAuth {
		// TODO: Implement authentication check
		logger.Debug("Page requires auth (not implemented)", map[string]interface{}{
			"page_slug": route.PageSlug,
		})
	}

	// Add route parameters to request context
	ctx := context.WithValue(r.Context(), "route_params", params)
	r = r.WithContext(ctx)

	// Render page
	logger.Debug("Rendering page", map[string]interface{}{
		"page_slug": route.PageSlug,
		"params":    params,
	})

	if err := sh.renderEngine.RenderPage(w, r, site, page); err != nil {
		logger.Error("Failed to render page", map[string]interface{}{
			"page_slug": route.PageSlug,
			"error":     err.Error(),
		})
		sh.renderError(w, r, site, logger, http.StatusInternalServerError, "Error rendering page", err.Error())
		return
	}

	logger.Debug("Page rendered successfully", map[string]interface{}{
		"page_slug": route.PageSlug,
	})
}

// populateRoutes loads all pages for a site and registers their routes
func (sh *SiteHandler) populateRoutes(site *common.Site, routeManager *common.RouteManager, logger *common.Logger) error {
	pageManager := common.NewPageManager(site)

	// Get all pages for the site (include unpublished for route setup)
	pages, err := pageManager.ListPages(true)
	if err != nil {
		return err
	}

	// Register each page's route
	for _, page := range pages {
		err := routeManager.AddRoute(page)
		if err != nil {
			logger.Error("Failed to add route", map[string]interface{}{
				"url":       page.Meta.URL,
				"page_slug": page.Slug,
				"error":     err.Error(),
			})
			continue
		}
		logger.Debug("Added route", map[string]interface{}{
			"url":       page.Meta.URL,
			"page_slug": page.Slug,
		})
	}

	return nil
}

// handleStaticAssets handles static asset serving
func (sh *SiteHandler) handleStaticAssets(w http.ResponseWriter, r *http.Request, site *common.Site, logger *common.Logger) {
	var assetPath string

	if strings.HasPrefix(r.URL.Path, "/assets/") {
		// Private assets
		assetPath = strings.TrimPrefix(r.URL.Path, "/assets/")
		logger.Debug("Serving private asset", map[string]interface{}{
			"asset_path": assetPath,
		})

		// Construct site-relative path for assets
		sitePath := filepath.Join("sites", site.Domain, "assets", assetPath)
		safeAssetPath, err := common.SecureServeFile(sitePath)
		if err != nil {
			logger.Warn("Asset access denied", map[string]interface{}{
				"path":  r.URL.Path,
				"error": err.Error(),
			})
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, safeAssetPath)
	} else if strings.HasPrefix(r.URL.Path, "/public/") {
		// Public assets
		assetPath = strings.TrimPrefix(r.URL.Path, "/public/")
		logger.Debug("Serving public asset", map[string]interface{}{
			"asset_path": assetPath,
		})

		// Construct site-relative path for public assets
		sitePath := filepath.Join("sites", site.Domain, "public", assetPath)
		safeAssetPath, err := common.SecureServeFile(sitePath)
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
func (sh *SiteHandler) handleAPI(w http.ResponseWriter, r *http.Request, site *common.Site, logger *common.Logger) {
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

// renderError renders an error page
func (sh *SiteHandler) renderError(w http.ResponseWriter, r *http.Request, site *common.Site, logger *common.Logger, statusCode int, message string, details ...string) {
	logger.Error("Rendering error page", map[string]interface{}{
		"status_code": statusCode,
		"message":     message,
		"path":        r.URL.Path,
		"details":     details,
	})
	if len(details) > 0 {
		message += ": " + strings.Join(details, ", ")
	}
	sh.renderEngine.RenderError(w, r, site, statusCode, message)
}
