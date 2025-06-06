package common

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

// RouteEntry represents a single route mapping
type RouteEntry struct {
	Pattern    string         // URL pattern like "/blog/post/:slug"
	PageSlug   string         // The page slug that handles this route
	Page       *Page          // The actual page object
	Parameters []string       // List of parameter names like ["slug"]
	Priority   int            // Lower number = higher priority
	Regex      *regexp.Regexp // Compiled regex for matching
}

// RouteManager manages dynamic routes for a site
type RouteManager struct {
	routes []RouteEntry
	router *chi.Mux
	site   *Site
	logger *Logger
	mu     sync.RWMutex
}

// NewRouteManager creates a new route manager for a site
func NewRouteManager(site *Site, logger *Logger) *RouteManager {
	return &RouteManager{
		routes: make([]RouteEntry, 0),
		router: chi.NewRouter(),
		site:   site,
		logger: logger,
	}
}

// parseRoutePattern converts a URL pattern like "/blog/post/:slug" into a regex
func (rm *RouteManager) parseRoutePattern(pattern string) (*regexp.Regexp, []string, error) {
	if pattern == "" {
		pattern = "/"
	}

	// Ensure pattern starts with /
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	// Track parameter names
	var parameters []string

	// Escape special regex characters except for our parameter syntax
	regexPattern := regexp.QuoteMeta(pattern)

	// Replace escaped parameter patterns with regex groups
	paramRegex := regexp.MustCompile(`\\:([a-zA-Z][a-zA-Z0-9_]*)\\:`)
	regexPattern = paramRegex.ReplaceAllStringFunc(regexPattern, func(match string) string {
		// Extract parameter name (remove the escaped colons)
		paramName := strings.Trim(match, "\\:")
		parameters = append(parameters, paramName)
		// Return regex pattern for the parameter
		return "([^/]+)"
	})

	// Anchor the pattern to match the full path
	regexPattern = "^" + regexPattern + "$"

	compiled, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile route pattern '%s': %w", pattern, err)
	}

	return compiled, parameters, nil
}

// calculatePriority determines the priority of a route pattern
// Lower numbers have higher priority
func (rm *RouteManager) calculatePriority(pattern string) int {
	if pattern == "/" {
		return 1000 // Home page gets lower priority to allow other exact matches
	}

	priority := 0
	segments := strings.Split(strings.Trim(pattern, "/"), "/")

	for _, segment := range segments {
		if strings.Contains(segment, ":") {
			// Parameter segments get lower priority
			priority += 100
		} else {
			// Static segments get higher priority
			priority += 1
		}
	}

	return priority
}

// AddRoute adds a new route for a page
func (rm *RouteManager) AddRoute(page *Page) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	pattern := page.Meta.URL
	if pattern == "" {
		// If no URL specified, use the page slug
		if page.Slug == "home" {
			pattern = "/"
		} else {
			pattern = "/" + page.Slug
		}
	}

	// Parse the route pattern
	regex, parameters, err := rm.parseRoutePattern(pattern)
	if err != nil {
		return fmt.Errorf("invalid route pattern for page '%s': %w", page.Slug, err)
	}

	// Calculate priority
	priority := rm.calculatePriority(pattern)

	// Create route entry
	entry := RouteEntry{
		Pattern:    pattern,
		PageSlug:   page.Slug,
		Page:       page,
		Parameters: parameters,
		Priority:   priority,
		Regex:      regex,
	}

	// Remove any existing route for this page
	rm.removeRouteForPage(page.Slug)

	// Add the new route
	rm.routes = append(rm.routes, entry)

	// Sort routes by priority (lower number = higher priority)
	sort.Slice(rm.routes, func(i, j int) bool {
		return rm.routes[i].Priority < rm.routes[j].Priority
	})

	rm.logger.Debug(fmt.Sprintf("Added route: %s -> %s (priority: %d)", pattern, page.Slug, priority))

	return nil
}

// removeRouteForPage removes the route for a specific page slug
func (rm *RouteManager) removeRouteForPage(pageSlug string) {
	for i := len(rm.routes) - 1; i >= 0; i-- {
		if rm.routes[i].PageSlug == pageSlug {
			rm.logger.Debug(fmt.Sprintf("Removed route: %s -> %s", rm.routes[i].Pattern, pageSlug))
			rm.routes = append(rm.routes[:i], rm.routes[i+1:]...)
			break
		}
	}
}

// RemoveRoute removes a route for a page
func (rm *RouteManager) RemoveRoute(pageSlug string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.removeRouteForPage(pageSlug)
}

// UpdateRoute updates the route for a page
func (rm *RouteManager) UpdateRoute(page *Page) error {
	return rm.AddRoute(page) // AddRoute already handles removing existing routes
}

// FindRoute finds the matching route for a given URL path
func (rm *RouteManager) FindRoute(urlPath string) (*RouteEntry, map[string]string, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Clean the URL path
	if urlPath == "" {
		urlPath = "/"
	}

	rm.logger.Debug(fmt.Sprintf("Looking for route matching: %s", urlPath))

	// Try to match against all routes in priority order
	for _, route := range rm.routes {
		matches := route.Regex.FindStringSubmatch(urlPath)
		if matches != nil {
			// Extract parameters
			params := make(map[string]string)
			for i, paramName := range route.Parameters {
				if i+1 < len(matches) {
					params[paramName] = matches[i+1]
				}
			}

			rm.logger.Debug(fmt.Sprintf("Route matched: %s -> %s with params: %v", urlPath, route.PageSlug, params))
			return &route, params, nil
		}
	}

	return nil, nil, fmt.Errorf("no route found for path: %s", urlPath)
}

// LoadAllRoutes loads routes for all pages in the site
func (rm *RouteManager) LoadAllRoutes() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Clear existing routes
	rm.routes = make([]RouteEntry, 0)

	pageManager := NewPageManager(rm.site)
	pages, err := pageManager.ListPages(true) // Include unpublished for route setup
	if err != nil {
		return fmt.Errorf("failed to list pages for route loading: %w", err)
	}

	rm.logger.Info(fmt.Sprintf("Loading routes for %d pages", len(pages)))

	for _, page := range pages {
		if err := rm.AddRoute(page); err != nil {
			rm.logger.Error(fmt.Sprintf("Failed to add route for page '%s': %v", page.Slug, err))
			continue
		}
	}

	rm.logger.Info(fmt.Sprintf("Loaded %d routes", len(rm.routes)))
	return nil
}

// GetAllRoutes returns all routes (for debugging/admin purposes)
func (rm *RouteManager) GetAllRoutes() []RouteEntry {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to prevent external modification
	routes := make([]RouteEntry, len(rm.routes))
	copy(routes, rm.routes)
	return routes
}

// RouteInfo provides information about a route for API responses
type RouteInfo struct {
	Pattern    string   `json:"pattern"`
	PageSlug   string   `json:"page_slug"`
	Parameters []string `json:"parameters"`
	Priority   int      `json:"priority"`
	PageTitle  string   `json:"page_title"`
	IsDraft    bool     `json:"is_draft"`
}

// GetRouteInfo returns route information for API responses
func (rm *RouteManager) GetRouteInfo() []RouteInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	info := make([]RouteInfo, len(rm.routes))
	for i, route := range rm.routes {
		info[i] = RouteInfo{
			Pattern:    route.Pattern,
			PageSlug:   route.PageSlug,
			Parameters: route.Parameters,
			Priority:   route.Priority,
			PageTitle:  route.Page.Meta.Title,
			IsDraft:    route.Page.Meta.IsDraft,
		}
	}

	return info
}

// Helper functions for request context
type contextKey string

const (
	routeParamsKey = contextKey("routeParams")
	matchedPageKey = contextKey("matchedPage")
)

// SetRouteParam sets a route parameter in the request context
func SetRouteParam(ctx context.Context, key, value string) context.Context {
	params, ok := ctx.Value(routeParamsKey).(map[string]string)
	if !ok {
		params = make(map[string]string)
	}
	params[key] = value
	return context.WithValue(ctx, routeParamsKey, params)
}

// SetMatchedPage sets the matched page in the request context
func SetMatchedPage(ctx context.Context, page *Page) context.Context {
	return context.WithValue(ctx, matchedPageKey, page)
}

// GetRouteParam gets a route parameter from the request context
func GetRouteParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(routeParamsKey).(map[string]string); ok {
		return params[key]
	}
	return ""
}

// GetMatchedPage gets the matched page from the request context
func GetMatchedPage(r *http.Request) *Page {
	if page, ok := r.Context().Value(matchedPageKey).(*Page); ok {
		return page
	}
	return nil
}
