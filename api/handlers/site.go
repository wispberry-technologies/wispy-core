package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/wispberry-technologies/wispy-core/common"
)

// SiteHandler handles requests for individual sites
type SiteHandler struct {
	siteManager  *common.SiteManager
	renderEngine *common.RenderEngine
}

// NewSiteHandler creates a new site handler
func NewSiteHandler(siteManager *common.SiteManager, renderEngine *common.RenderEngine) *SiteHandler {
	return &SiteHandler{
		siteManager:  siteManager,
		renderEngine: renderEngine,
	}
}

// ServeHTTP handles HTTP requests for sites
func (sh *SiteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Debug: Request Host: %s, Path: %s", r.Host, r.URL.Path)

	// Get site from host header
	site, err := sh.siteManager.GetSiteFromHost(r.Host)
	if err != nil {
		log.Printf("Debug: Failed to get site: %v", err)
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	log.Printf("Debug: Found site: %s", site.Domain)

	// Check if site is active
	if !site.IsActive {
		http.Error(w, "Site is not active", http.StatusServiceUnavailable)
		return
	}

	// Handle static assets
	if strings.HasPrefix(r.URL.Path, "/assets/") || strings.HasPrefix(r.URL.Path, "/public/") {
		sh.handleStaticAssets(w, r, site)
		return
	}

	// Handle API requests
	if strings.HasPrefix(r.URL.Path, "/api/") {
		sh.handleAPI(w, r, site)
		return
	}

	// Handle page requests
	sh.handlePage(w, r, site)
}

// handlePage handles page rendering
func (sh *SiteHandler) handlePage(w http.ResponseWriter, r *http.Request, site *common.Site) {
	// Extract page slug from URL path
	slug := strings.Trim(r.URL.Path, "/")
	log.Printf("Debug: Page slug: '%s'", slug)

	// Create page manager
	pageManager := common.NewPageManager(site)

	// Get page
	page, err := pageManager.GetPage(slug)
	if err != nil {
		log.Printf("Debug: Failed to get page: %v", err)
		sh.renderError(w, r, site, http.StatusNotFound, "Page not found")
		return
	}

	log.Printf("Debug: Found page: %s", page.Meta.Title)

	// Check if page is draft and user is not authenticated admin
	if page.Meta.IsDraft {
		// TODO: Implement admin authentication check
		// For now, we'll show drafts in development
	}

	// Check if page requires authentication
	if page.Meta.RequireAuth {
		// TODO: Implement authentication check
		// For now, we'll skip authentication
	}

	// Render page
	log.Printf("Debug: Rendering page...")
	if err := sh.renderEngine.RenderPage(w, r, site, page); err != nil {
		log.Printf("Debug: Failed to render page: %v", err)
		sh.renderError(w, r, site, http.StatusInternalServerError, "Error rendering page")
		return
	}
	log.Printf("Debug: Page rendered successfully")
}

// handleStaticAssets handles static asset serving
func (sh *SiteHandler) handleStaticAssets(w http.ResponseWriter, r *http.Request, site *common.Site) {
	var assetPath string

	if strings.HasPrefix(r.URL.Path, "/assets/") {
		// Private assets
		assetPath = strings.TrimPrefix(r.URL.Path, "/assets/")
		http.ServeFile(w, r, site.AssetsPath+"/"+assetPath)
	} else if strings.HasPrefix(r.URL.Path, "/public/") {
		// Public assets
		assetPath = strings.TrimPrefix(r.URL.Path, "/public/")
		http.ServeFile(w, r, site.PublicPath+"/"+assetPath)
	} else {
		http.NotFound(w, r)
	}
}

// handleAPI handles API requests
func (sh *SiteHandler) handleAPI(w http.ResponseWriter, r *http.Request, site *common.Site) {
	// TODO: Implement API handling
	// For now, return a simple JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "API not implemented yet"}`))
}

// renderError renders an error page
func (sh *SiteHandler) renderError(w http.ResponseWriter, r *http.Request, site *common.Site, statusCode int, message string) {
	sh.renderEngine.RenderError(w, r, site, statusCode, message)
}
