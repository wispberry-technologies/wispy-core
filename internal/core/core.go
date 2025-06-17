// Package core provides the core functionality for the CMS
package core

import (
	"os"
	"path/filepath"
	"sync"

	"wispy-core/internal/cache"
	"wispy-core/internal/core/parser"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"
)

// ImportAllSites loads all sites from the given directory
func ImportAllSites(sitesPath string) (map[string]*models.SiteInstance, error) {
	sites := make(map[string]*models.SiteInstance)

	// List the site directories
	entries, err := filepath.Glob(filepath.Join(sitesPath, "*"))
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// Check if this is a directory
		info, err := os.Stat(entry)
		if err != nil || !info.IsDir() {
			continue
		}

		// Get the domain name from the directory name
		domain := filepath.Base(entry)

		// Create a new site instance
		site := NewSiteInstance(domain)
		// Load the pages for this site
		if err := parser.LoadPagesForSite(site); err != nil {
			common.Warning("Error loading pages for site %s: %v", domain, err)
			continue
		}

		// Add the site to the map
		sites[domain] = site
	}

	return sites, nil
}

func NewSiteInstance(domain string) *models.SiteInstance {
	siteInstance := &models.SiteInstance{
		Domain:   domain,
		Name:     domain,
		BasePath: common.RootSitesPath(domain),
		IsActive: true,
		Theme:    "default",
		Config: models.SiteConfig{
			CssProcessor: "wispy-tail",
		},
		DBCache:        cache.NewDBCache(),
		SecurityConfig: &models.SiteSecurityConfig{},
		Templates:      make(map[string]string),
		Pages:          make(map[string]*models.Page), // routes for this site
		Mu:             sync.RWMutex{},                // mutex for thread-safe route access
		DBManager:      nil,                           // Will be initialized separately
	}
	return siteInstance
}
