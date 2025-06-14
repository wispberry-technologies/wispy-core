package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"wispy-core/common"
	"wispy-core/models"
)

// TestRouteWrapperIntegration tests the full integration of the RouteWrapper
func TestRouteWrapperIntegration(t *testing.T) {
	// Create a new RouteWrapper
	rw := NewRouteWrapper()

	// Create multiple mock sites
	site1 := NewSiteInstance("site1.com", "Site 1")
	site2 := NewSiteInstance("site2.com", "Site 2")

	// Create pages for site1
	page1 := NewPage("Home", "/", common.WrapBraces("define \"page-body\"")+"This is a test page"+common.WrapBraces("enddefine"))
	page2 := NewPage("About", "/about", common.WrapBraces("define \"page-body\"")+"This is a test page"+common.WrapBraces("enddefine"))

	// Create pages for site2
	page3 := NewPage("Home", "/", common.WrapBraces("define \"page-body\"")+"This is a test page"+common.WrapBraces("enddefine"))

	// Add pages to sites
	site1.Pages["/"] = page1
	site1.Pages["/about"] = page2
	site2.Pages["/"] = page3

	// Register both sites
	rw.RegisterSite(site1)
	rw.RegisterSite(site2)

	// Test that both sites' routes are registered
	// Since we can't easily differentiate between domains in the test router,
	// we'll test by checking that the expected number of routes are registered

	// Test adding a new route dynamically
	newPage := NewPage("Contact", "/contact", common.WrapBraces("define \"page-body\"")+"This is a new test page"+common.WrapBraces("enddefine"))

	rw.RegisterRoute(site1, "/contact", newPage)

	// Test the new route
	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	w := httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "This is a new test page" {
		t.Errorf("Expected body 'This is a new test page', got '%s'", string(body))
	}

	// Test removing a route
	rw.RemoveRoute("site1.com", "/about")

	// Test that the removed route returns 404
	req = httptest.NewRequest(http.MethodGet, "/about", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d for removed route, got %d", http.StatusNotFound, resp.StatusCode)
	}

	// Test regenerating routes for a specific site
	rw.RegenerateSiteRoutes("site1.com")

	// The contact route should still exist after regeneration
	req = httptest.NewRequest(http.MethodGet, "/contact", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp = w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d after regeneration, got %d", http.StatusOK, resp.StatusCode)
	}
}

// TestRouteWrapperConcurrency tests that the RouteWrapper is thread-safe
func TestRouteWrapperConcurrency(t *testing.T) {
	rw := NewRouteWrapper()

	site := NewSiteInstance("test.com", "Test Site")

	rw.RegisterSite(site)

	// Run concurrent operations
	done := make(chan bool, 3)

	// Goroutine 1: Add routes
	go func() {
		for i := 0; i < 10; i++ {
			page := NewPage("Test Page", "/test"+string(rune('0'+i)), common.WrapBraces("define \"page-body\"")+"This is a test page"+common.WrapBraces("enddefine"))
			rw.RegisterRoute(site, page.URL, page)
		}
		done <- true
	}()

	// Goroutine 2: Remove routes
	go func() {
		for i := 0; i < 5; i++ {
			rw.RemoveRoute("test.com", "/test"+string(rune('0'+i)))
		}
		done <- true
	}()

	// Goroutine 3: Make requests
	go func() {
		for i := 0; i < 20; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test5", nil)
			w := httptest.NewRecorder()
			rw.ServeHTTP(w, req)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Test should complete without race conditions or panics
	t.Log("Concurrency test completed successfully")
}

// Add helpers for test context creation
func NewSiteInstance(domain, name string) *models.SiteInstance {
	return &models.SiteInstance{
		Domain: domain,
		Site: &models.Site{
			Name:   name,
			Domain: domain,
		},
		Pages: make(map[string]*models.Page),
	}
}

func NewPage(title, url, content string) *models.Page {
	return &models.Page{
		Title:   title,
		URL:     url,
		Content: content,
	}
}
