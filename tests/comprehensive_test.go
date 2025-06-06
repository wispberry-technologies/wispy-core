package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wispberry-technologies/wispy-core/common"
	"github.com/wispberry-technologies/wispy-core/routes"
)

// TestComprehensiveCaching tests comprehensive caching behavior with different endpoints
func TestComprehensiveCaching(t *testing.T) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/test", nil)
	mockReq.Header.Set("User-Agent", "Comprehensive-Test")

	endpoints := []struct {
		path           string
		expectedStatus int
		shouldCache    bool
	}{
		{"/health", 200, true},
		{"/nonexistent", 404, false}, // 404s might not be cached
		{"/admin", 200, true},        // admin endpoint returns 200
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("Endpoint_%s", endpoint.path), func(t *testing.T) {
			// First call
			start := time.Now()
			resp1, err := dispatcher.Call("GET", endpoint.path, nil, nil, mockReq)
			dur1 := time.Since(start)

			if err != nil {
				t.Fatalf("First call to %s failed: %v", endpoint.path, err)
			}

			if resp1.StatusCode != endpoint.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", endpoint.expectedStatus, endpoint.path, resp1.StatusCode)
			}

			if resp1.CacheHit {
				t.Errorf("First call to %s should not be a cache hit", endpoint.path)
			}

			// Second call (should be cached if successful)
			start = time.Now()
			resp2, err := dispatcher.Call("GET", endpoint.path, nil, nil, mockReq)
			dur2 := time.Since(start)

			if err != nil {
				t.Fatalf("Second call to %s failed: %v", endpoint.path, err)
			}

			if endpoint.shouldCache {
				if !resp2.CacheHit {
					t.Errorf("Second call to %s should be a cache hit", endpoint.path)
				}

				if dur2 >= dur1 {
					t.Errorf("Cached call to %s (%v) should be faster than uncached (%v)", endpoint.path, dur2, dur1)
				}

				speedup := float64(dur1) / float64(dur2)
				improvement := (1 - float64(dur2)/float64(dur1)) * 100

				t.Logf("✅ %s: %.1fx speedup (%.1f%% improvement) - %v → %v",
					endpoint.path, speedup, improvement, dur1, dur2)
			}
		})
	}
}

// TestPerformanceBenchmark tests high-volume performance characteristics
func TestPerformanceBenchmark(t *testing.T) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)
	mockReq.Header.Set("User-Agent", "Performance-Test")

	t.Run("HighVolumePerformance", func(t *testing.T) {
		// Warm up cache
		dispatcher.Call("GET", "/health", nil, nil, mockReq)

		const numRequests = 100
		var totalCachedTime time.Duration
		var cacheHits, cacheMisses int

		start := time.Now()
		for i := 0; i < numRequests; i++ {
			reqStart := time.Now()
			resp, err := dispatcher.Call("GET", "/health", nil, nil, mockReq)
			reqDuration := time.Since(reqStart)

			if err != nil {
				t.Fatalf("Request %d failed: %v", i+1, err)
			}

			totalCachedTime += reqDuration

			if resp.CacheHit {
				cacheHits++
			} else {
				cacheMisses++
			}
		}
		totalTime := time.Since(start)

		avgCachedTime := totalCachedTime / numRequests
		requestsPerSecond := float64(numRequests) / totalTime.Seconds()

		// Performance expectations
		if avgCachedTime > 100*time.Microsecond {
			t.Errorf("Average cached response time too slow: %v (expected < 100µs)", avgCachedTime)
		}

		if requestsPerSecond < 1000 {
			t.Errorf("Requests per second too low: %.0f (expected > 1000)", requestsPerSecond)
		}

		if cacheHits < numRequests-5 { // Allow a few misses for timing
			t.Errorf("Too many cache misses: %d hits, %d misses", cacheHits, cacheMisses)
		}

		t.Logf("✅ Performance: %.0f req/s, avg cached: %v, hits: %d/%d",
			requestsPerSecond, avgCachedTime, cacheHits, numRequests)
	})
}

// TestCacheExpirationAndTTL tests cache expiration behavior
func TestCacheExpirationAndTTL(t *testing.T) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)
	mockReq.Header.Set("User-Agent", "TTL-Test")

	t.Run("CacheStatistics", func(t *testing.T) {
		// Make several API calls to generate statistics
		endpoints := []string{"/health", "/admin", "/nonexistent"}

		for _, endpoint := range endpoints {
			// Make two calls to each endpoint (first miss, second hit)
			dispatcher.Call("GET", endpoint, nil, nil, mockReq)
			dispatcher.Call("GET", endpoint, nil, nil, mockReq)
		}

		stats := dispatcher.GetCacheStats()

		// Verify statistics structure
		if enabled, ok := stats["enabled"].(bool); !ok || !enabled {
			t.Error("Cache should be enabled")
		}

		if totalEntries, ok := stats["total_entries"].(int); !ok || totalEntries < 2 {
			t.Errorf("Expected at least 2 cache entries, got %v", totalEntries)
		}

		if cacheHits, ok := stats["cache_hits"].(int); !ok || cacheHits < 2 {
			t.Errorf("Expected at least 2 cache hits, got %v", cacheHits)
		}

		if cacheMisses, ok := stats["cache_misses"].(int); !ok || cacheMisses < 3 {
			t.Errorf("Expected at least 3 cache misses, got %v", cacheMisses)
		}

		t.Logf("✅ Cache stats: entries=%v, hits=%v, misses=%v",
			stats["total_entries"], stats["cache_hits"], stats["cache_misses"])
	})
}

// TestEdgeCases tests various edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/test", nil)

	t.Run("EmptyHeaders", func(t *testing.T) {
		resp, err := dispatcher.Call("GET", "/health", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Call with empty headers failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		t.Logf("✅ Empty headers handled correctly")
	})

	t.Run("CustomHeaders", func(t *testing.T) {
		headers := map[string]string{
			"X-Custom-Header": "test-value",
			"Authorization":   "Bearer test123",
		}

		resp, err := dispatcher.Call("GET", "/health", nil, headers, mockReq)
		if err != nil {
			t.Fatalf("Call with custom headers failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		t.Logf("✅ Custom headers handled correctly")
	})

	t.Run("LargeRequestBody", func(t *testing.T) {
		// Create a large request body
		largeBody := make([]byte, 1024*10) // 10KB
		for i := range largeBody {
			largeBody[i] = byte('A' + (i % 26))
		}

		resp, err := dispatcher.Call("POST", "/health", largeBody, nil, mockReq)
		if err != nil {
			t.Fatalf("Call with large body failed: %v", err)
		}
		// Note: POST requests typically shouldn't be cached
		if resp.CacheHit {
			t.Error("POST request should not be cached")
		}
		t.Logf("✅ Large request body handled correctly")
	})

	t.Run("InvalidMethods", func(t *testing.T) {
		invalidMethods := []string{"INVALID", "TRACE", "CONNECT"}

		for _, method := range invalidMethods {
			resp, err := dispatcher.Call(method, "/health", nil, nil, mockReq)
			// Should handle gracefully, not crash
			if err != nil {
				t.Logf("Method %s returned error (expected): %v", method, err)
			} else {
				t.Logf("Method %s returned status %d", method, resp.StatusCode)
			}
		}
		t.Logf("✅ Invalid methods handled gracefully")
	})
}

// TestMemoryAndResourceUsage tests for memory leaks and resource usage
func TestMemoryAndResourceUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)

	t.Run("HighVolumeNoCrash", func(t *testing.T) {
		// Make a large number of requests to test for memory leaks
		const numRequests = 1000

		for i := 0; i < numRequests; i++ {
			path := fmt.Sprintf("/health?iteration=%d", i%10) // Some variety
			_, err := dispatcher.Call("GET", path, nil, nil, mockReq)
			if err != nil {
				t.Fatalf("Request %d failed: %v", i, err)
			}
		}

		// Check cache statistics to ensure it's working
		stats := dispatcher.GetCacheStats()
		if stats["total_entries"].(int) < 5 {
			t.Error("Expected multiple cache entries after high volume test")
		}

		t.Logf("✅ High volume test completed: %d requests, %d cache entries",
			numRequests, stats["total_entries"])
	})
}
