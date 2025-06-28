package site

import (
	"net/http"
	"path/filepath"

	"wispy-core/common"
)

// ScaffoldSiteRoutes sets up the basic routes for a site
func ScaffoldSiteRoutes(s Site) {
	router := s.GetRouter()

	// Setup routes for the site
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Welcome to " + s.GetName()))
	})

	// Add more route handlers as needed
	router.Get("/about", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("About " + s.GetName()))
	})

	// Setup static file routes for site assets
	router.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		// Get the path from the URL
		path := r.URL.Path[len("/public/"):]
		// Serve the file from the site's assets directory
		http.ServeFile(w, r, filepath.Join(s.GetContentDir(), "public", path))
	})

	common.Info("Scaffolded routes for site: %s", s.GetName())
}

// ScaffoldAllSites sets up routes for all sites
func ScaffoldAllSites(sites map[string]Site) {
	for _, s := range sites {
		ScaffoldSiteRoutes(s)
	}
}
