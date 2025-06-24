// Package core provides the core functionality for the CMS
package core

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wispy-core/internal/cache"
	"wispy-core/internal/core/pages"
	"wispy-core/internal/core/parser"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

func NewSiteInstance(domain string) *models.SiteInstance {
	basePath := common.RootSitesPath(domain)
	siteInstance := &models.SiteInstance{
		Domain:   domain,
		Name:     domain,
		BasePath: basePath,
		IsActive: true,
		Theme:    "default",
		DBCache:  cache.NewDBCache(),
		Router:   chi.NewRouter(),
		AuthConfig: &models.SiteAuthConfig{
			// Security settings
			MaxFailedLoginAttempts:         5,
			FailedLoginAttemptLockDuration: 30 * time.Minute,
			// Registration settings
			RegistrationEnabled:  true,
			RequiredFields:       []string{"display_name", "email", "password"},
			DefaultRoles:         []string{},
			AllowedPasswordReset: true,
		},
		RouteProxies:   make(map[string]string),
		OAuthProviders: []string{},
		CssProcessor:   "wispy-tail",
		// Templates:      make(map[string]string),
		Pages: make(map[string]*models.Page),
		Mu:    sync.RWMutex{},
	}
	LoadSiteConfig(basePath, siteInstance)

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
	common.Info("Loading Pages for - " + siteInstance.Domain)

	// Pull existing pages from Database
	db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "pages")
	if err != nil {
		common.Error("failed to load pages db for " + siteInstance.Domain)
	}

	var hasRetried = false
	// re-try point
QueryPages:
	pageRows, listErr := pages.ListPages(db, 100, 0)
	if listErr != nil {
		common.Warning("could not Query pages for " + siteInstance.Domain + ": " + listErr.Error())
		if !hasRetried {
			common.Info("attempting to scaffold pages db for " + siteInstance.Domain)
			hasRetried = true
			pages.ScaffoldPagesDb(db)
			LoadPagesFromFiles(siteInstance, db)
			goto QueryPages
		} else {
			return fmt.Errorf("Failed to load any pages for " + siteInstance.Domain)
		}
	}

	common.Info("Loaded %d pages from database for %s", len(pageRows), siteInstance.Domain)
	// Add database pages to site instance
	for _, page := range pageRows {
		siteInstance.Pages[page.Slug] = page
	}

	return nil
}

func LoadPagesFromFiles(siteInstance *models.SiteInstance, db *sql.DB) error {
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
		common.Info("Processing page: %s", page.Slug)
		_, err = pages.InsertPage(db, page)
		if err != nil {
			common.Error("Failed to insert page %s: %v", page.Slug, err)
			return nil
		}
		return nil
	})
	return err
}

// StaticAndSitePublicHandler returns an http.HandlerFunc that serves static files from a site's public directory and a global static directory.
// It checks for directory traversal attempts and serves files from the appropriate paths.
func StaticAndSitePublicHandler(sitesPath, staticPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Basic security check - prevent directory traversal
		if strings.Contains(path, "..") {
			common.Warning("Forbidden access attempt to path: %s", path)
			common.Warning("- Request from IP: %s", r.RemoteAddr)
			common.Warning("- Request headers: %v", r.Header)

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		//
		domain := common.NormalizeHost(r.Host)
		publicPath := filepath.Join(sitesPath, domain)
		staticPath := filepath.Join(staticPath)
		//
		switch {
		case strings.HasPrefix(path, "/public/"):
			filePath := filepath.Join(publicPath, path)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				common.Warning("File not found: %s", filePath)
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, filePath)
		case strings.HasPrefix(path, "/static/"):
			filePath := filepath.Join(staticPath, path)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				common.Warning("File not found: %s", filePath)
				http.NotFound(w, r)
				return
			}
			http.ServeFile(w, r, filePath)
		//
		default:
			common.Warning("Invalid static path: %s", path)
			http.Error(w, "Invalid static path", http.StatusBadRequest)
			return
		}
		// 404
		// If we reach here, it means the file was not found
		common.Warning("File not found: %s", path)
		http.NotFound(w, r)
		// Log the request details
		common.Info("- Request from IP: %s", r.RemoteAddr)
		common.Info("- Request headers: %v", r.Header)
		//
	}
}
