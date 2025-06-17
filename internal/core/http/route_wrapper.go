// Package http provides HTTP handler functions and routing functionality for the CMS
package http

import (
	"net/http"
	"sync"
	"time"

	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

// RouteInfo provides information about a registered route
type RouteInfo struct {
	Domain    string
	URL       string
	PageName  string
	Layout    string
	HookCount int
}

// RouteStats tracks statistics for routes
type RouteStats struct {
	RequestCount    int64         `json:"request_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastAccessed    time.Time     `json:"last_accessed"`
}

// RouteWrapper is a wrapper around chi.Router that allows for dynamic route regeneration
type RouteWrapper struct {
	mu           sync.RWMutex
	router       chi.Router
	sites        map[string]*models.SiteInstance
	routeConfigs map[string]map[string]RouteConfiguration // domain -> url -> config
	routeStats   map[string]*RouteStats                   // url -> stats
	enableStats  bool                                     // Whether to collect route statistics
}

// RouteConfiguration stores configuration for a specific route
type RouteConfiguration struct {
	Page     *models.Page
	Instance *models.SiteInstance
	URL      string
	Hooks    []func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool
}

// NewRouteWrapperWithStats creates a new RouteWrapper with optional statistics
func NewRouteWrapperWithStats(router chi.Router, enableStats bool) *RouteWrapper {
	return &RouteWrapper{
		router:       router,
		sites:        make(map[string]*models.SiteInstance),
		routeConfigs: make(map[string]map[string]RouteConfiguration),
		routeStats:   make(map[string]*RouteStats),
		enableStats:  enableStats,
	}
}

// RegisterSite registers a site with the RouteWrapper
func (rw *RouteWrapper) RegisterSite(site *models.SiteInstance) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.sites[site.Domain] = site
	if _, exists := rw.routeConfigs[site.Domain]; !exists {
		rw.routeConfigs[site.Domain] = make(map[string]RouteConfiguration)
	}

	// Register routes for all pages
	for _, page := range site.Pages {
		// Skip draft pages in production
		if page.IsDraft {
			continue
		}

		urlPath := "/" + page.Slug
		if site.Domain != "localhost" {
			urlPath = "/" + site.Domain + urlPath
		}

		// Store route config
		rw.routeConfigs[site.Domain][urlPath] = RouteConfiguration{
			Page:     page,
			Instance: site,
			URL:      urlPath,
			Hooks:    []func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool{},
		}

		// // Register the route with Chi
		// rw.router.Get(urlPath, func(w http.ResponseWriter, r *http.Request) {
		// 	// Get the site instance
		// 	domain := site.Domain
		// 	if domain == "" {
		// 		domain = "localhost"
		// 	}

		// 	// Get the route configuration
		// 	rw.mu.RLock()
		// 	routeConfig, exists := rw.routeConfigs[domain][urlPath]
		// 	rw.mu.RUnlock()

		// 	if !exists {
		// 		http.Error(w, "Page not found", http.StatusNotFound)
		// 		return
		// 	}

		// 	// Execute route handler
		// 	rw.handleRoute(w, r, routeConfig)
		// })
	}
}

// ServeHTTP implements the http.Handler interface by delegating to the underlying chi router
func (rw *RouteWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw.mu.RLock()
	router := rw.router
	rw.mu.RUnlock()

	// Create a simple wrapper to track statistics
	if rw.enableStats {
		start := time.Now()
		url := r.URL.Path
		defer func() {
			duration := time.Since(start)
			rw.mu.Lock()
			stats, exists := rw.routeStats[url]
			if !exists {
				stats = &RouteStats{
					LastAccessed: start,
				}
				rw.routeStats[url] = stats
			}
			stats.LastAccessed = start
			stats.TotalDuration += duration
			stats.RequestCount++
			stats.AverageDuration = time.Duration(stats.TotalDuration.Nanoseconds() / stats.RequestCount)
			rw.mu.Unlock()
		}()
	}

	router.ServeHTTP(w, r)
}
