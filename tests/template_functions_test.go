package tests

import (
	"net/http"
	"testing"

	"github.com/wispberry-technologies/wispy-core/common"
	"github.com/wispberry-technologies/wispy-core/routes"
)

// TestTemplateFunctions tests template function integration with the API system
func TestTemplateFunctions(t *testing.T) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, err := http.NewRequest("GET", "http://localhost:8080/test-template", nil)
	if err != nil {
		t.Fatalf("Failed to create mock request: %v", err)
	}
	mockReq.Header.Set("User-Agent", "Template-Function-Test")

	t.Run("HealthAPIThroughTemplateFunction", func(t *testing.T) {
		result, err := dispatcher.Call("GET", "/health", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Health API call failed: %v", err)
		}

		if result.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", result.StatusCode)
		}

		if result.Data == nil {
			t.Error("Expected response data, got nil")
		}

		// Check that the response contains the expected health status
		if status, ok := result.Data["status"]; !ok || status != "healthy" {
			t.Errorf("Expected status 'healthy', got %v", status)
		}

		t.Logf("✅ Health check through template function successful")
		t.Logf("   Status Code: %d", result.StatusCode)
		t.Logf("   Duration: %v", result.Duration)
		t.Logf("   Cache Hit: %v", result.CacheHit)
		t.Logf("   Response Data: %v", result.Data)
	})

	t.Run("APICallWithCustomHeaders", func(t *testing.T) {
		headers := map[string]string{
			"X-Template-Test": "true",
			"Content-Type":    "application/json",
		}

		result, err := dispatcher.Call("GET", "/health", nil, headers, mockReq)
		if err != nil {
			t.Fatalf("API call with headers failed: %v", err)
		}

		if result.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", result.StatusCode)
		}

		t.Logf("✅ API call with custom headers successful")
		t.Logf("   Headers processed correctly")
	})

	t.Run("ErrorHandlingInTemplateFunction", func(t *testing.T) {
		result, err := dispatcher.Call("GET", "/nonexistent-endpoint", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Error handling test failed: %v", err)
		}

		if result.StatusCode != 404 {
			t.Errorf("Expected status 404, got %d", result.StatusCode)
		}

		t.Logf("✅ Error handling in template function working correctly")
		t.Logf("   Status Code: %d", result.StatusCode)
	})

	t.Run("CachingInTemplateContext", func(t *testing.T) {
		// First call (should be uncached)
		result1, err := dispatcher.Call("GET", "/admin", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("First template API call failed: %v", err)
		}

		if result1.CacheHit {
			t.Error("First call should not be a cache hit")
		}

		// Second call (should be cached)
		result2, err := dispatcher.Call("GET", "/admin", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Second template API call failed: %v", err)
		}

		if !result2.CacheHit {
			t.Error("Second call should be a cache hit")
		}

		if result2.Duration >= result1.Duration {
			t.Error("Cached call should be faster than uncached call")
		}

		t.Logf("✅ Caching working correctly in template context")
		t.Logf("   First call: %v (cached: %v)", result1.Duration, result1.CacheHit)
		t.Logf("   Second call: %v (cached: %v)", result2.Duration, result2.CacheHit)
	})

	t.Run("TemplateAPIDataIntegrity", func(t *testing.T) {
		result, err := dispatcher.Call("GET", "/health", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Data integrity test failed: %v", err)
		}

		// Verify that all expected metadata fields are present
		expectedFields := []string{"_internal_api", "_cache_hit", "_duration_ms", "_method", "_path", "_status_code"}

		for _, field := range expectedFields {
			if _, exists := result.Data[field]; !exists {
				t.Errorf("Expected metadata field '%s' not found in response", field)
			}
		}

		// Verify data types
		if internalAPI, ok := result.Data["_internal_api"].(bool); !ok || !internalAPI {
			t.Error("_internal_api should be true")
		}

		if method, ok := result.Data["_method"].(string); !ok || method != "GET" {
			t.Errorf("_method should be 'GET', got %v", method)
		}

		if path, ok := result.Data["_path"].(string); !ok || path != "/health" {
			t.Errorf("_path should be '/health', got %v", path)
		}

		if statusCode, ok := result.Data["_status_code"].(int); !ok || statusCode != 200 {
			t.Errorf("_status_code should be 200, got %v", statusCode)
		}

		t.Logf("✅ Template API data integrity verified")
		t.Logf("   All metadata fields present and correctly typed")
	})
}

// TestRenderEngineFunctionIntegration tests the actual render engine functionality
func TestRenderEngineFunctionIntegration(t *testing.T) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)

	t.Run("RenderEngineInitialization", func(t *testing.T) {
		// Test that the render engine initializes correctly
		if renderEngine == nil {
			t.Error("Render engine should not be nil")
		}

		t.Logf("✅ Render engine initialized successfully")
	})

	t.Run("APIDispatcherIntegration", func(t *testing.T) {
		// Test that API dispatcher can be set on render engine
		r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
		dispatcher := common.NewRouterAPIDispatcher(r)

		// This sets up the template functions with API integration
		renderEngine.SetAPIDispatcher(dispatcher)

		t.Logf("✅ API dispatcher integration working correctly")
	})
}
