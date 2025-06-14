package core

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"wispy-core/common"
	"wispy-core/models"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
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

// MiddlewareFunc is a function that can be applied to all routes
type MiddlewareFunc func(http.Handler) http.Handler

// RouteWrapper is a wrapper around chi.Router that allows for dynamic route regeneration
type RouteWrapper struct {
	mu           sync.RWMutex
	router       chi.Router
	sites        map[string]*models.SiteInstance
	routeConfigs map[string]map[string]RouteConfig // domain -> url -> config
	middlewares  []MiddlewareFunc                  // Global middlewares
	routeStats   map[string]*RouteStats            // url -> stats
	enableStats  bool                              // Whether to collect route statistics
}

// RouteConfig stores configuration for a specific route
type RouteConfig struct {
	Page  *models.Page
	Site  *models.SiteInstance
	URL   string
	Hooks []func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool
}

// NewRouteWrapper creates a new RouteWrapper instance
func NewRouteWrapper() *RouteWrapper {
	return &RouteWrapper{
		router:       chi.NewRouter(),
		sites:        make(map[string]*models.SiteInstance),
		routeConfigs: make(map[string]map[string]RouteConfig),
		middlewares:  []MiddlewareFunc{},
		routeStats:   make(map[string]*RouteStats),
		enableStats:  true,
	}
}

// NewRouteWrapperWithStats creates a new RouteWrapper with optional statistics
func NewRouteWrapperWithStats(enableStats bool) *RouteWrapper {
	return &RouteWrapper{
		router:       chi.NewRouter(),
		sites:        make(map[string]*models.SiteInstance),
		routeConfigs: make(map[string]map[string]RouteConfig),
		middlewares:  []MiddlewareFunc{},
		routeStats:   make(map[string]*RouteStats),
		enableStats:  enableStats,
	}
}

// ServeHTTP implements the http.Handler interface by delegating to the underlying chi router
func (rw *RouteWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw.mu.RLock()
	router := rw.router
	rw.mu.RUnlock()

	router.ServeHTTP(w, r)
}

// RegisterSite adds all pages from a site to the router
func (rw *RouteWrapper) RegisterSite(site *models.SiteInstance) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Store the site
	rw.sites[site.Domain] = site

	// Initialize route configs for this domain if not already present
	if _, exists := rw.routeConfigs[site.Domain]; !exists {
		rw.routeConfigs[site.Domain] = make(map[string]RouteConfig)
	}

	// Add all pages from the site
	for url, page := range site.Pages {
		config := RouteConfig{
			Page:  page,
			Site:  site,
			URL:   url,
			Hooks: []func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool{},
		}

		rw.routeConfigs[site.Domain][url] = config
	}

	// Regenerate all routes
	rw.regenerateAllRoutes()
}

// RegisterRoute adds a single page/route to the router
func (rw *RouteWrapper) RegisterRoute(site *models.SiteInstance, url string, page *models.Page) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Store the site if not already present
	if _, exists := rw.sites[site.Domain]; !exists {
		rw.sites[site.Domain] = site
	}

	// Initialize route configs for this domain if not already present
	if _, exists := rw.routeConfigs[site.Domain]; !exists {
		rw.routeConfigs[site.Domain] = make(map[string]RouteConfig)
	}

	// Add the route config
	config := RouteConfig{
		Page:  page,
		Site:  site,
		URL:   url,
		Hooks: []func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool{},
	}

	rw.routeConfigs[site.Domain][url] = config

	// Regenerate all routes
	rw.regenerateAllRoutes()
}

// RemoveRoute removes a single route from the router
func (rw *RouteWrapper) RemoveRoute(domain, url string) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Remove the route config if it exists
	if domainConfigs, domainExists := rw.routeConfigs[domain]; domainExists {
		if _, routeExists := domainConfigs[url]; routeExists {
			delete(rw.routeConfigs[domain], url)

			// Regenerate all routes
			rw.regenerateAllRoutes()
		}
	}
}

// AddRouteHook adds a hook function to a specific route
// The hook should return true to continue processing the request or false to abort
func (rw *RouteWrapper) AddRouteHook(domain, url string, hook func(http.ResponseWriter, *http.Request, *models.Page, *models.SiteInstance) bool) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if domainConfigs, domainExists := rw.routeConfigs[domain]; domainExists {
		if config, routeExists := domainConfigs[url]; routeExists {
			config.Hooks = append(config.Hooks, hook)
			rw.routeConfigs[domain][url] = config

			// Regenerate all routes
			rw.regenerateAllRoutes()
		}
	}
}

// RegenerateSiteRoutes regenerates all routes for a specific site
func (rw *RouteWrapper) RegenerateSiteRoutes(domain string) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Create a new router
	newRouter := chi.NewRouter()

	// Add all routes from all sites
	for siteDomain, siteConfigs := range rw.routeConfigs {
		if siteDomain == domain || domain == "" {
			for _, config := range siteConfigs {
				rw.addRouteToRouter(newRouter, config)
			}
		}
	}

	// Replace the old router with the new one
	rw.router = newRouter
}

// regenerateAllRoutes recreates all routes in the chi router
func (rw *RouteWrapper) regenerateAllRoutes() {
	// Create a new router
	newRouter := chi.NewRouter()

	// Apply global middlewares
	for _, middleware := range rw.middlewares {
		newRouter.Use(middleware)
	}

	// Add all routes from all sites
	for _, siteConfigs := range rw.routeConfigs {
		for _, config := range siteConfigs {
			rw.addRouteToRouter(newRouter, config)
		}
	}

	// Replace the old router with the new one
	rw.router = newRouter
}

// addRouteToRouter adds a single route to the given router
func (rw *RouteWrapper) addRouteToRouter(router chi.Router, config RouteConfig) {
	router.Get(config.URL, func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Run hooks before processing the page
		for _, hook := range config.Hooks {
			if !hook(w, r, config.Page, config.Site) {
				return // Abort processing if a hook returns false
			}
		}

		// Create template context using NewTemplateContext
		data := map[string]interface{}{"Page": config.Page, "Site": config.Site.Site}
		engine := NewTemplateEngine(DefaultFunctionMap())
		ctx := NewTemplateContext(data, engine)
		ctx.Request = r

		// Determine layout to use
		layoutName := config.Page.Layout
		if layoutName == "" {
			layoutName = "default"
		}

		layoutAsString := GetLayout(config.Site.Domain, layoutName)
		result, errs := engine.Render(config.Page.Content+layoutAsString, ctx)
		for _, err := range errs {
			log.Printf("[ERROR]Render: Url(%s): %v", config.URL, err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(result))

		dur := time.Since(start)
		log.Printf("[INFO] Rendered %s in %s", config.URL, dur)

		// Update route statistics
		if rw.enableStats {
			rw.updateRouteStats(config.URL, dur)
		}
	})
}

// updateRouteStats updates the statistics for a specific route
func (rw *RouteWrapper) updateRouteStats(url string, duration time.Duration) {
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

// GetRouter returns the underlying chi.Router
func (rw *RouteWrapper) GetRouter() chi.Router {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.router
}

// AddDefaultMiddlewares adds default middlewares to the router
func (rw *RouteWrapper) AddDefaultMiddlewares() {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	// Add default middlewares
	rw.router.Use(middleware.Logger)
	rw.router.Use(middleware.Recoverer)
	rw.router.Use(middleware.Timeout(60 * time.Second))
	// Setup rate-limiters
	rw.router.Use(httprate.Limit(
		requestsPerSecond, // requests
		1*time.Second,     // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))
	rw.router.Use(httprate.LimitByIP(requestsPerMinute, time.Minute))

}

// AddGlobalMiddleware adds a middleware that will be applied to all routes
func (rw *RouteWrapper) AddGlobalMiddleware(middleware MiddlewareFunc) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	rw.middlewares = append(rw.middlewares, middleware)
	rw.regenerateAllRoutes()
}

// ListRoutes returns information about all registered routes
func (rw *RouteWrapper) ListRoutes() []RouteInfo {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	var routes []RouteInfo
	for domain, domainConfigs := range rw.routeConfigs {
		for url, config := range domainConfigs {
			routes = append(routes, RouteInfo{
				Domain:    domain,
				URL:       url,
				PageName:  config.Page.Title, // Using Title since there's no Name field
				Layout:    config.Page.Layout,
				HookCount: len(config.Hooks),
			})
		}
	}
	return routes
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
		"status":           "healthy",
		"total_routes":     rw.GetRouteCount(),
		"total_sites":      len(rw.sites),
		"middleware_count": len(rw.middlewares),
		"timestamp":        time.Now().UTC(),
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
func (rw *RouteWrapper) GetRoute(domain, url string) (*RouteConfig, bool) {
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
			rw.regenerateAllRoutes()
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
	rw.regenerateAllRoutes()
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
		result["name"] = site.Site.Name
		result["is_active"] = site.Site.IsActive
		result["theme"] = site.Site.Theme
		result["route_count"] = len(rw.routeConfigs[domain])
		result["has_templates"] = len(site.Templates) > 0
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
