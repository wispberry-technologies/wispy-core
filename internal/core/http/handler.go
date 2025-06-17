// Package http provides HTTP handler functions for the CMS
package http

import (
	"log"
	"net/http"

	"wispy-core/internal/core/html"
	"wispy-core/pkg/models"
)

// RegisterRoutes registers all routes for the given site instances - simplified version
func RegisterRoutes(mux *http.ServeMux, siteInstances map[string]*models.SiteInstance) {
	for domain, site := range siteInstances {
		log.Printf("Registering routes for domain: %s", domain)

		// Register routes for each page in the site
		for slug, page := range site.Pages {
			// Create a closure to capture the current page
			currentPage := page
			currentSite := site

			pattern := "/" + slug
			if slug == "home" || slug == "" {
				pattern = "/"
			}

			mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
				handlePageRequest(w, r, currentPage, currentSite)
			})
		}
	}
}

// handlePageRequest handles requests for individual pages
func handlePageRequest(w http.ResponseWriter, r *http.Request, page *models.Page, site *models.SiteInstance) {
	// Use the existing render function
	data := make(map[string]interface{})
	html.RenderPageWithLayout(w, r, page, site, data)
}
