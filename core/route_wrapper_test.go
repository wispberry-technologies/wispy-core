package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"wispy-core/common"
	"wispy-core/models"
)

func TestRouteWrapper(t *testing.T) {
	// Create a new RouteWrapper
	rw := NewRouteWrapper()

	// Create a mock site instance
	mockSite := &models.SiteInstance{
		Domain: "example.com",
		Site: &models.Site{
			Name:   "Test Site",
			Domain: "example.com",
		},
		Pages: make(map[string]*models.Page),
	}

	// Create a mock page
	mockPage := &models.Page{
		Title:   "Test Page",
		URL:     "/test",
		Content: common.WrapBraces("define \"page-body\"") + "This is a test page" + common.WrapBraces("enddefine"),
	}

	// Add the page to the site
	mockSite.Pages["/test"] = mockPage

	// Register the site with the route wrapper
	rw.RegisterSite(mockSite)

	// Test the route
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	// Check that the response contains the expected content
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "This is a test page" {
		t.Errorf("Expected body 'This is a test page', got '%s'", string(body))
	}

	// Test route hooks
	hookCalled := false
	rw.AddRouteHook("example.com", "/test", func(w http.ResponseWriter, r *http.Request, page *models.Page, site *models.SiteInstance) bool {
		hookCalled = true
		return true
	})

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	if !hookCalled {
		t.Error("Hook was not called")
	}

	// Test adding a new route
	newPage := &models.Page{
		Title:   "New Test Page",
		URL:     "/new-test",
		Content: common.WrapBraces("define \"page-body\"") + "This is a new test page" + common.WrapBraces("enddefine"),
	}

	rw.RegisterRoute(mockSite, "/new-test", newPage)

	req = httptest.NewRequest(http.MethodGet, "/new-test", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if string(body) != "This is a new test page" {
		t.Errorf("Expected body 'This is a new test page', got '%s'", string(body))
	}

	// Test removing a route
	rw.RemoveRoute("example.com", "/test")

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp = w.Result()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	// Test hook that aborts request
	rw.AddRouteHook("example.com", "/new-test", func(w http.ResponseWriter, r *http.Request, page *models.Page, site *models.SiteInstance) bool {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied"))
		return false
	})

	req = httptest.NewRequest(http.MethodGet, "/new-test", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, resp.StatusCode)
	}

	if string(body) != "Access denied" {
		t.Errorf("Expected body 'Access denied', got '%s'", string(body))
	}

	// Test regenerating routes for a specific site
	rw.RegenerateSiteRoutes("example.com")

	req = httptest.NewRequest(http.MethodGet, "/new-test", nil)
	w = httptest.NewRecorder()

	rw.ServeHTTP(w, req)

	resp = w.Result()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}
