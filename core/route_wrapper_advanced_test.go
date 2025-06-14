package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wispy-core/common"
	"wispy-core/models"
)

func TestRouteWrapperUtilityMethods(t *testing.T) {
	rw := NewRouteWrapper()

	// Create test site and page
	site := &models.SiteInstance{
		Domain: "test.example.com",
		Site: &models.Site{
			Domain:   "test.example.com",
			Name:     "Test Site",
			IsActive: true,
			Theme:    "default",
		},
		Pages: make(map[string]*models.Page),
	}

	page := &models.Page{
		Title:   "Test Page",
		Layout:  "default",
		URL:     "/test",
		Content: common.WrapBraces("define \"page-body\"") + "Test content" + common.WrapBraces("enddefine"),
	}

	// Test RegisterRoute
	rw.RegisterRoute(site, "/test", page)

	// Test RouteExists
	if !rw.RouteExists("test.example.com", "/test") {
		t.Error("Expected route to exist")
	}

	if rw.RouteExists("test.example.com", "/nonexistent") {
		t.Error("Expected route to not exist")
	}

	// Test GetRoute
	config, exists := rw.GetRoute("test.example.com", "/test")
	if !exists {
		t.Error("Expected route config to exist")
	}
	if config.Page.Title != "Test Page" {
		t.Errorf("Expected page title 'Test Page', got '%s'", config.Page.Title)
	}

	// Test GetSiteDomains
	domains := rw.GetSiteDomains()
	if len(domains) != 1 || domains[0] != "test.example.com" {
		t.Errorf("Expected domains ['test.example.com'], got %v", domains)
	}

	// Test GetSiteInfo
	siteInfo := rw.GetSiteInfo("test.example.com")
	if !siteInfo["exists"].(bool) {
		t.Error("Expected site to exist")
	}
	if siteInfo["name"] != "Test Site" {
		t.Errorf("Expected site name 'Test Site', got '%v'", siteInfo["name"])
	}

	// Test UpdatePageInRoute
	newPage := &models.Page{
		Title:   "Updated Test Page",
		Layout:  "default",
		URL:     "/test",
		Content: common.WrapBraces("define \"page-body\"") + "Updated content" + common.WrapBraces("enddefine"),
	}

	if !rw.UpdatePageInRoute("test.example.com", "/test", newPage) {
		t.Error("Expected page update to succeed")
	}

	config, _ = rw.GetRoute("test.example.com", "/test")
	if config.Page.Title != "Updated Test Page" {
		t.Errorf("Expected updated page title 'Updated Test Page', got '%s'", config.Page.Title)
	}

	// Test RemoveSite
	rw.RemoveSite("test.example.com")
	if rw.RouteExists("test.example.com", "/test") {
		t.Error("Expected route to be removed after site removal")
	}

	domains = rw.GetSiteDomains()
	if len(domains) != 0 {
		t.Errorf("Expected no domains after site removal, got %v", domains)
	}
}

func TestRouteWrapperValidation(t *testing.T) {
	rw := NewRouteWrapper()

	// Create two sites with conflicting URLs
	site1 := &models.SiteInstance{
		Domain: "site1.com",
		Site:   &models.Site{Domain: "site1.com"},
		Pages:  make(map[string]*models.Page),
	}

	site2 := &models.SiteInstance{
		Domain: "site2.com",
		Site:   &models.Site{Domain: "site2.com"},
		Pages:  make(map[string]*models.Page),
	}

	page1 := &models.Page{Title: "Page 1", URL: "/conflict", Content: common.WrapBraces("define \"page-body\"") + "Content 1" + common.WrapBraces("enddefine")}
	page2 := &models.Page{Title: "Page 2", URL: "/conflict", Content: common.WrapBraces("define \"page-body\"") + "Content 2" + common.WrapBraces("enddefine")}

	rw.RegisterRoute(site1, "/conflict", page1)
	rw.RegisterRoute(site2, "/conflict", page2)

	// Test ValidateRoutes
	issues := rw.ValidateRoutes()
	if len(issues) == 0 {
		t.Error("Expected validation issues for conflicting URLs")
	}

	if len(issues) > 0 && !contains(issues[0], "URL conflict") {
		t.Errorf("Expected URL conflict issue, got %s", issues[0])
	}
}

func TestRouteWrapperHealthCheck(t *testing.T) {
	rw := NewRouteWrapper()

	// Test health check on empty wrapper
	health := rw.HealthCheck()
	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
	if health["total_routes"] != 0 {
		t.Errorf("Expected 0 routes, got %v", health["total_routes"])
	}

	// Add some routes
	site := &models.SiteInstance{
		Domain: "health.example.com",
		Site:   &models.Site{Domain: "health.example.com"},
		Pages:  make(map[string]*models.Page),
	}

	page := &models.Page{Title: "Health Page", URL: "/health", Content: common.WrapBraces("define \"page-body\"") + "Healthy" + common.WrapBraces("enddefine")}
	rw.RegisterRoute(site, "/health", page)

	health = rw.HealthCheck()
	if health["total_routes"] != 1 {
		t.Errorf("Expected 1 route, got %v", health["total_routes"])
	}
	if health["total_sites"] != 1 {
		t.Errorf("Expected 1 site, got %v", health["total_sites"])
	}

	// Check timestamp
	if _, ok := health["timestamp"].(time.Time); !ok {
		t.Error("Expected timestamp to be a time.Time")
	}
}

func TestRouteWrapperMiddleware(t *testing.T) {
	rw := NewRouteWrapper()

	// Test middleware
	middlewareExecuted := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareExecuted = true
			w.Header().Set("X-Middleware", "executed")
			next.ServeHTTP(w, r)
		})
	}

	rw.AddGlobalMiddleware(middleware)

	// Add a test route
	site := &models.SiteInstance{
		Domain: "middleware.example.com",
		Site:   &models.Site{Domain: "middleware.example.com"},
		Pages:  make(map[string]*models.Page),
	}

	page := &models.Page{
		Title:   "Middleware Test",
		URL:     "/middleware",
		Content: common.WrapBraces("define \"page-body\"") + "This is a test page" + common.WrapBraces("enddefine"),
	}

	rw.RegisterRoute(site, "/middleware", page)

	// Test the route
	req := httptest.NewRequest("GET", "/middleware", nil)
	req.Host = "middleware.example.com"
	w := httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	if !middlewareExecuted {
		t.Error("Expected middleware to be executed")
	}

	if w.Header().Get("X-Middleware") != "executed" {
		t.Error("Expected middleware header to be set")
	}

	// Check health to verify middleware count
	health := rw.HealthCheck()
	if health["middleware_count"] != 1 {
		t.Errorf("Expected 1 middleware, got %v", health["middleware_count"])
	}
}

func TestRouteWrapperStatistics(t *testing.T) {
	rw := NewRouteWrapper()

	// Create test site and page
	site := &models.SiteInstance{
		Domain: "stats.example.com",
		Site:   &models.Site{Domain: "stats.example.com"},
		Pages:  make(map[string]*models.Page),
	}

	page := &models.Page{
		Title:   "Stats Test",
		URL:     "/stats",
		Content: common.WrapBraces("define \"page-body\"") + "This is a test page" + common.WrapBraces("enddefine"),
	}

	rw.RegisterRoute(site, "/stats", page)

	// Test that stats are enabled by default
	if !rw.IsStatsEnabled() {
		t.Error("Expected stats to be enabled by default")
	}

	// Make a request to generate stats
	req := httptest.NewRequest("GET", "/stats", nil)
	req.Host = "stats.example.com"
	w := httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	// Check that stats were recorded
	stats, exists := rw.GetRouteStats("/stats")
	if !exists {
		t.Error("Expected route stats to exist")
	}

	if stats.RequestCount != 1 {
		t.Errorf("Expected request count 1, got %d", stats.RequestCount)
	}

	if stats.TotalDuration <= 0 {
		t.Error("Expected total duration to be greater than 0")
	}

	if stats.AverageDuration <= 0 {
		t.Error("Expected average duration to be greater than 0")
	}

	// Make another request
	rw.ServeHTTP(w, req)

	stats, _ = rw.GetRouteStats("/stats")
	if stats.RequestCount != 2 {
		t.Errorf("Expected request count 2, got %d", stats.RequestCount)
	}

	// Test GetAllRouteStats
	allStats := rw.GetAllRouteStats()
	if len(allStats) != 1 {
		t.Errorf("Expected 1 route in stats, got %d", len(allStats))
	}

	if allStats["/stats"].RequestCount != 2 {
		t.Errorf("Expected request count 2 in all stats, got %d", allStats["/stats"].RequestCount)
	}

	// Test GetTopRoutes
	topRoutes := rw.GetTopRoutes(5)
	if len(topRoutes) != 1 {
		t.Errorf("Expected 1 top route, got %d", len(topRoutes))
	}

	if topRoutes[0].URL != "/stats" {
		t.Errorf("Expected top route URL '/stats', got '%s'", topRoutes[0].URL)
	}

	if topRoutes[0].RequestCount != 2 {
		t.Errorf("Expected top route request count 2, got %d", topRoutes[0].RequestCount)
	}

	// Test ResetRouteStats
	rw.ResetRouteStats("/stats")
	stats, _ = rw.GetRouteStats("/stats")
	if stats.RequestCount != 0 {
		t.Errorf("Expected request count 0 after reset, got %d", stats.RequestCount)
	}

	// Test disabling stats
	rw.EnableStats(false)
	if rw.IsStatsEnabled() {
		t.Error("Expected stats to be disabled")
	}

	// Make a request with stats disabled
	rw.ServeHTTP(w, req)
	stats, _ = rw.GetRouteStats("/stats")
	if stats.RequestCount != 0 {
		t.Errorf("Expected request count to remain 0 with stats disabled, got %d", stats.RequestCount)
	}

	// Re-enable stats
	rw.EnableStats(true)
	if !rw.IsStatsEnabled() {
		t.Error("Expected stats to be enabled")
	}
}

func TestRouteWrapperStatsWithMultipleRoutes(t *testing.T) {
	rw := NewRouteWrapper()

	// Create test site with multiple pages
	site := &models.SiteInstance{
		Domain: "multi.example.com",
		Site:   &models.Site{Domain: "multi.example.com"},
		Pages:  make(map[string]*models.Page),
	}

	pages := []*models.Page{
		{Title: "Home", URL: "/", Content: common.WrapBraces("define \"page-body\"") + "This is a test page" + common.WrapBraces("enddefine")},
		{Title: "About", URL: "/about", Content: common.WrapBraces("define \"page-body\"") + "This is a test page" + common.WrapBraces("enddefine")},
		{Title: "Contact", URL: "/contact", Content: common.WrapBraces("define \"page-body\"") + "This is a test page" + common.WrapBraces("enddefine")},
	}

	// Register routes
	for _, page := range pages {
		rw.RegisterRoute(site, page.URL, page)
	}

	// Make different numbers of requests to each route
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Host = "multi.example.com"
	req2 := httptest.NewRequest("GET", "/about", nil)
	req2.Host = "multi.example.com"
	req3 := httptest.NewRequest("GET", "/contact", nil)
	req3.Host = "multi.example.com"
	w := httptest.NewRecorder()

	// 3 requests to home
	rw.ServeHTTP(w, req1)
	rw.ServeHTTP(w, req1)
	rw.ServeHTTP(w, req1)

	// 2 requests to about
	rw.ServeHTTP(w, req2)
	rw.ServeHTTP(w, req2)

	// 1 request to contact
	rw.ServeHTTP(w, req3)

	// Test GetTopRoutes
	topRoutes := rw.GetTopRoutes(0) // No limit
	if len(topRoutes) != 3 {
		t.Errorf("Expected 3 top routes, got %d", len(topRoutes))
	}

	// Should be sorted by request count (descending)
	if topRoutes[0].URL != "/" || topRoutes[0].RequestCount != 3 {
		t.Errorf("Expected first route to be '/' with 3 requests, got '%s' with %d", topRoutes[0].URL, topRoutes[0].RequestCount)
	}

	if topRoutes[1].URL != "/about" || topRoutes[1].RequestCount != 2 {
		t.Errorf("Expected second route to be '/about' with 2 requests, got '%s' with %d", topRoutes[1].URL, topRoutes[1].RequestCount)
	}

	if topRoutes[2].URL != "/contact" || topRoutes[2].RequestCount != 1 {
		t.Errorf("Expected third route to be '/contact' with 1 request, got '%s' with %d", topRoutes[2].URL, topRoutes[2].RequestCount)
	}

	// Test with limit
	topRoutes = rw.GetTopRoutes(2)
	if len(topRoutes) != 2 {
		t.Errorf("Expected 2 top routes with limit, got %d", len(topRoutes))
	}

	// Test ResetAllRouteStats
	rw.ResetAllRouteStats()
	allStats := rw.GetAllRouteStats()
	for url, stats := range allStats {
		if stats.RequestCount != 0 {
			t.Errorf("Expected request count 0 for %s after reset all, got %d", url, stats.RequestCount)
		}
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || contains(s[1:], substr)))
}
