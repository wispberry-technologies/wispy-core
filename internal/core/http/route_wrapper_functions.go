package http

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
	"wispy-core/internal/core/html"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

// regenerateAllRoutes recreates all routes in the chi router
func (rw *RouteWrapper) RegenerateAllRoutes() {
	// Create a new router
	newRouter := chi.NewRouter()

	// Add static file serving for each site's public directory
	for domain, site := range rw.sites {
		rw.AddStaticFileServing(newRouter, domain, site)
	}

	// Add all routes from all sites
	for _, siteConfigs := range rw.routeConfigs {
		for _, config := range siteConfigs {
			rw.AddRouteToRouter(newRouter, config)
		}
	}

	// Add 404 handler
	// newRouter.NotFound(rw.handle404)

	// Replace the old router with the new one
	rw.router = newRouter
}

// addRouteToRouter adds a single route to the given router
func (rw *RouteWrapper) AddRouteToRouter(router chi.Router, route RouteConfiguration) {
	router.Get(route.URL, func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Run hooks before processing the page
		for _, hook := range route.Hooks {
			if !hook(w, r, route.Page, route.Instance) {
				return // Abort processing if a hook returns false
			}
		}

		dur := time.Since(start)
		log.Printf("Rendered %s in %s", route.URL, dur)

		html.RenderPageWithLayout(w, r, route.Page, route.Instance, map[string]interface{}{
			"Site": route.Page.SiteDetails,
		})

		// Update route statistics
		if rw.enableStats {
			rw.UpdateRouteStats(route.URL, dur)
		}
	})
}

// AddRouteHook adds a hook function to be called before rendering a page
func (rw *RouteWrapper) AddRouteHook(domain, url string, hook func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if _, exists := rw.routeConfigs[domain]; !exists {
		return
	}

	if routeConfig, exists := rw.routeConfigs[domain][url]; exists {
		routeConfig.Hooks = append(routeConfig.Hooks, hook)
		rw.routeConfigs[domain][url] = routeConfig
	}
}

// addStaticFileServing adds static file serving for a site's public and assets directories
func (rw *RouteWrapper) AddStaticFileServing(router chi.Router, domain string, site *models.SiteInstance) {
	// Serve files from /public/
	publicPath := common.RootSitesPath(domain) + "/public/"
	router.Get("/public/*", func(w http.ResponseWriter, r *http.Request) {
		// Strip the /public prefix from the URL path
		path := strings.TrimPrefix(r.URL.Path, "/public/")
		filePath := publicPath + path

		// Basic security check - prevent directory traversal
		if strings.Contains(path, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		http.ServeFile(w, r, filePath)
	})

	// Serve files from /assets/
	assetsPath := common.RootSitesPath(domain) + "/assets/"
	router.Get("/assets/*", func(w http.ResponseWriter, r *http.Request) {
		// Strip the /assets prefix from the URL path
		path := strings.TrimPrefix(r.URL.Path, "/assets/")
		filePath := assetsPath + path

		// Basic security check - prevent directory traversal
		if strings.Contains(path, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		http.ServeFile(w, r, filePath)
	})
}

// GetRegisteredRoutes returns information about all registered routes
func (rw *RouteWrapper) GetRegisteredRoutes() []RouteInfo {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	var routes []RouteInfo
	for domain, configs := range rw.routeConfigs {
		for url, config := range configs {
			routes = append(routes, RouteInfo{
				Domain:    domain,
				URL:       url,
				PageName:  config.Page.Title,
				Layout:    config.Page.LayoutName,
				HookCount: len(config.Hooks),
			})
		}
	}
	return routes
}

// GetRouter returns the underlying chi router
func (rw *RouteWrapper) GetRouter() chi.Router {
	rw.mu.RLock()
	defer rw.mu.RUnlock()
	return rw.router
}

// updateRouteStats updates the statistics for a specific route
func (rw *RouteWrapper) UpdateRouteStats(url string, duration time.Duration) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	stats, exists := rw.routeStats[url]
	if !exists {
		stats = &RouteStats{}
		rw.routeStats[url] = stats
	}

	atomic.AddInt64(&stats.RequestCount, 1)
	stats.TotalDuration += duration
	stats.AverageDuration = stats.TotalDuration / time.Duration(stats.RequestCount)
	stats.LastAccessed = time.Now()
}

// GetRouteCount returns the total number of registered routes
func (rw *RouteWrapper) GetRouteCount() int {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	count := 0
	for _, domainConfigs := range rw.routeConfigs {
		count += len(domainConfigs)
	}
	return count
}

// GetSiteRouteCount returns the number of routes for a specific site
func (rw *RouteWrapper) GetSiteRouteCount(domain string) int {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if domainConfigs, exists := rw.routeConfigs[domain]; exists {
		return len(domainConfigs)
	}
	return 0
}

// ValidateRoutes checks for route conflicts and returns any issues found
func (rw *RouteWrapper) ValidateRoutes() []string {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	var issues []string
	urlCounts := make(map[string][]string) // url -> domains that use it

	for domain, domainConfigs := range rw.routeConfigs {
		for url := range domainConfigs {
			urlCounts[url] = append(urlCounts[url], domain)
		}
	}

	// Check for URL conflicts across domains
	for url, domains := range urlCounts {
		if len(domains) > 1 {
			issues = append(issues, fmt.Sprintf("URL conflict: %s used by domains: %v", url, domains))
		}
	}

	return issues
}

// HealthCheck performs a basic health check on the RouteWrapper
func (rw *RouteWrapper) HealthCheck() map[string]interface{} {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	result := map[string]interface{}{
		"status":       "healthy",
		"total_routes": rw.GetRouteCount(),
		"total_sites":  len(rw.sites),
		"timestamp":    time.Now().UTC(),
	}

	// Check for issues
	issues := rw.ValidateRoutes()
	if len(issues) > 0 {
		result["status"] = "warning"
		result["issues"] = issues
	}

	return result
}

// GetRoute returns the configuration for a specific route
func (rw *RouteWrapper) GetRoute(domain, url string) (*RouteConfiguration, bool) {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if domainConfigs, exists := rw.routeConfigs[domain]; exists {
		if config, routeExists := domainConfigs[url]; routeExists {
			return &config, true
		}
	}
	return nil, false
}

// UpdatePageInRoute updates the page content for an existing route
func (rw *RouteWrapper) UpdatePageInRoute(domain, url string, page *models.Page) bool {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if domainConfigs, exists := rw.routeConfigs[domain]; exists {
		if config, routeExists := domainConfigs[url]; routeExists {
			config.Page = page
			rw.routeConfigs[domain][url] = config
			rw.RegenerateAllRoutes()
			return true
		}
	}
	return false
}

// RemoveSite removes all routes for a specific site
func (rw *RouteWrapper) RemoveSite(domain string) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Remove all route configs for this domain
	delete(rw.routeConfigs, domain)

	// Remove the site instance
	delete(rw.sites, domain)

	// Regenerate all routes
	rw.RegenerateAllRoutes()
}

// GetSiteDomains returns a list of all registered site domains
func (rw *RouteWrapper) GetSiteDomains() []string {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	domains := make([]string, 0, len(rw.sites))
	for domain := range rw.sites {
		domains = append(domains, domain)
	}
	return domains
}

// GetSiteInfo returns basic information about a specific site
func (rw *RouteWrapper) GetSiteInfo(domain string) map[string]interface{} {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	result := map[string]interface{}{
		"domain": domain,
		"exists": false,
	}

	if site, exists := rw.sites[domain]; exists {
		result["exists"] = true
		result["name"] = site.Name
		result["is_active"] = site.IsActive
		result["theme"] = site.Theme
		result["route_count"] = len(rw.routeConfigs[domain])
	}

	return result
}

// RouteExists checks if a specific route exists
func (rw *RouteWrapper) RouteExists(domain, url string) bool {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if domainConfigs, exists := rw.routeConfigs[domain]; exists {
		_, routeExists := domainConfigs[url]
		return routeExists
	}
	return false
}

// GetRouteStats returns statistics for a specific route
func (rw *RouteWrapper) GetRouteStats(url string) (*RouteStats, bool) {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	if stats, exists := rw.routeStats[url]; exists {
		// Return a copy to avoid race conditions
		return &RouteStats{
			RequestCount:    atomic.LoadInt64(&stats.RequestCount),
			TotalDuration:   stats.TotalDuration,
			AverageDuration: stats.AverageDuration,
			LastAccessed:    stats.LastAccessed,
		}, true
	}
	return nil, false
}

// GetAllRouteStats returns statistics for all routes
func (rw *RouteWrapper) GetAllRouteStats() map[string]RouteStats {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	result := make(map[string]RouteStats)
	for url, stats := range rw.routeStats {
		result[url] = RouteStats{
			RequestCount:    atomic.LoadInt64(&stats.RequestCount),
			TotalDuration:   stats.TotalDuration,
			AverageDuration: stats.AverageDuration,
			LastAccessed:    stats.LastAccessed,
		}
	}
	return result
}

// ResetRouteStats resets statistics for a specific route
func (rw *RouteWrapper) ResetRouteStats(url string) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if stats, exists := rw.routeStats[url]; exists {
		atomic.StoreInt64(&stats.RequestCount, 0)
		stats.TotalDuration = 0
		stats.AverageDuration = 0
		stats.LastAccessed = time.Time{}
	}
}

// ResetAllRouteStats resets statistics for all routes
func (rw *RouteWrapper) ResetAllRouteStats() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	for _, stats := range rw.routeStats {
		atomic.StoreInt64(&stats.RequestCount, 0)
		stats.TotalDuration = 0
		stats.AverageDuration = 0
		stats.LastAccessed = time.Time{}
	}
}

// EnableStats enables or disables route statistics collection
func (rw *RouteWrapper) EnableStats(enable bool) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.enableStats = enable
}

// IsStatsEnabled returns whether statistics collection is enabled
func (rw *RouteWrapper) IsStatsEnabled() bool {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.enableStats
}

// GetTopRoutes returns the most requested routes
func (rw *RouteWrapper) GetTopRoutes(limit int) []struct {
	URL          string
	RequestCount int64
} {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	type routeCount struct {
		URL          string
		RequestCount int64
	}

	var routes []routeCount
	for url, stats := range rw.routeStats {
		routes = append(routes, routeCount{
			URL:          url,
			RequestCount: atomic.LoadInt64(&stats.RequestCount),
		})
	}

	// Simple sort by request count (descending)
	for i := 0; i < len(routes)-1; i++ {
		for j := 0; j < len(routes)-i-1; j++ {
			if routes[j].RequestCount < routes[j+1].RequestCount {
				routes[j], routes[j+1] = routes[j+1], routes[j]
			}
		}
	}

	if limit > 0 && limit < len(routes) {
		routes = routes[:limit]
	}

	result := make([]struct {
		URL          string
		RequestCount int64
	}, len(routes))

	for i, route := range routes {
		result[i] = struct {
			URL          string
			RequestCount int64
		}{
			URL:          route.URL,
			RequestCount: route.RequestCount,
		}
	}

	return result
}
