// Package core provides the core functionality for the CMS
package core

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wispy-core/internal/cache"
	"wispy-core/internal/core/parser"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

func NewSiteInstance(domain string) *models.SiteInstance {
	siteInstance := &models.SiteInstance{
		Domain:   domain,
		Name:     domain,
		BasePath: common.RootSitesPath(domain),
		IsActive: true,
		Theme:    "default",
		Config: models.SiteConfig{
			OAuthProviders: []string{"discord", "github"}, // default OAuth providers
			CssProcessor:   "wispy-tail",
		},
		DBCache: cache.NewDBCache(),
		Router:  chi.NewRouter(),
		SecurityConfig: &models.SiteSecurityConfig{
			MaxFailedLoginAttempts:         5,
			FailedLoginAttemptLockDuration: 30 * time.Minute,
		},
		Templates: make(map[string]string),
		Pages:     make(map[string]*models.Page), // routes for this site
		Mu:        sync.RWMutex{},                // mutex for thread-safe route access
	}
	return siteInstance
}

// LoadAllSites loads all sites from the given directory
func LoadAllSitesAsInstances(sitesPath string) (map[string]*models.SiteInstance, error) {
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
		if err := LoadPagesForSite(site); err != nil {
			common.Warning("Error loading pages for site %s: %v", domain, err)
			continue
		}

		// Add the site to the map
		sites[domain] = site
	}

	return sites, nil
}

// LoadPagesForSite loads all pages for a given site instance
// Takes explicit site instance and returns a map of pages
func LoadPagesForSite(siteInstance *models.SiteInstance) error {
	pagesDir := common.RootSitesPath(siteInstance.Domain, "pages")
	err := filepath.WalkDir(pagesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errored files/dirs
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".html" {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			return nil
		}
		page, err := parser.ParsePageHTML(siteInstance, string(content))
		if err != nil {
			common.Warning("Error parsing page %s: %v", d.Name(), err)
			return nil
		}
		// should be safe since we are using filepath.WalkDir on a path that is guaranteed to be within the site instance's pages directory
		pathParts := strings.SplitN(path, "pages", 2)
		pageFilePath := pathParts[1]
		page.FilePath = strings.TrimPrefix(pageFilePath, "/") // remove leading slash

		// Store the page in the SiteInstance's Pages map, keyed by Slug
		common.Info("page: %s", page.Slug)
		siteInstance.Pages[page.Slug] = page
		return nil
	})
	return err
}

// StaticFileServingWithoutContextHandler returns an http.HandlerFunc that serves static files from a site's public directory without site context
func StaticFileServingWithoutContextHandler(sitesPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := chi.URLParam(r, "domain")
		publicPath := filepath.Join(sitesPath, domain, "public")

		// Strip the /public/{domain} prefix from the URL path
		path := strings.TrimPrefix(r.URL.Path, "/public/"+domain+"/")
		filePath := filepath.Join(publicPath, path)

		// Basic security check - prevent directory traversal
		if strings.Contains(path, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, filePath)
	}
}
