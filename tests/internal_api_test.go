package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/wispberry-technologies/wispy-core/common"
	"github.com/wispberry-technologies/wispy-core/routes"
)

// setupTestEnvironment creates a test environment for internal API testing
func setupTestEnvironment(t *testing.T) (*common.RouterAPIDispatcher, *http.Request) {
	t.Helper()

	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, err := http.NewRequest("GET", "http://localhost:8080/test", nil)
	if err != nil {
		t.Fatalf("Failed to create mock request: %v", err)
	}
	mockReq.Header.Set("User-Agent", "Test-Suite")

	return dispatcher, mockReq
}

// TestBasicAPIFunctionality tests basic API functionality
func TestBasicAPIFunctionality(t *testing.T) {
	dispatcher, mockReq := setupTestEnvironment(t)

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := dispatcher.Call("GET", "/health", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if resp.Data == nil {
			t.Error("Expected response data, got nil")
		}

		if status, ok := resp.Data["status"]; !ok || status != "healthy" {
			t.Errorf("Expected status 'healthy', got %v", status)
		}

		t.Logf("✅ Health check successful: %v (duration: %v)", resp.Data["status"], resp.Duration)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		resp, err := dispatcher.Call("GET", "/nonexistent", nil, nil, mockReq)
		if err != nil {
			t.Fatalf("Error handling test failed: %v", err)
		}

		if resp.StatusCode != 404 {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		t.Logf("✅ Error handling working: status %d", resp.StatusCode)
	})
}

// TestCachingPerformance tests caching performance improvements
func TestCachingPerformance(t *testing.T) {
	dispatcher, mockReq := setupTestEnvironment(t)

	t.Run("CacheSpeedup", func(t *testing.T) {
		// First call (uncached)
		start := time.Now()
		resp1, err := dispatcher.Call("GET", "/admin", nil, nil, mockReq)
		uncachedTime := time.Since(start)

		if err != nil {
			t.Fatalf("First API call failed: %v", err)
		}

		if resp1.CacheHit {
			t.Error("First call should not be a cache hit")
		}

		// Second call (should be cached)
		start = time.Now()
		resp2, err := dispatcher.Call("GET", "/admin", nil, nil, mockReq)
		cachedTime := time.Since(start)

		if err != nil {
			t.Fatalf("Second API call failed: %v", err)
		}

		if !resp2.CacheHit {
			t.Error("Second call should be a cache hit")
		}

		// Calculate performance improvement
		if cachedTime >= uncachedTime {
			t.Error("Cached call should be faster than uncached call")
		}

		speedup := float64(uncachedTime) / float64(cachedTime)
		improvement := (1 - float64(cachedTime)/float64(uncachedTime)) * 100

		if improvement < 50 { // Expect at least 50% improvement
			t.Errorf("Expected significant performance improvement, got %.1f%%", improvement)
		}

		t.Logf("✅ Cache performance: %.1fx speedup (%.1f%% improvement)", speedup, improvement)
		t.Logf("   Uncached: %v → Cached: %v", uncachedTime, cachedTime)
	})

	t.Run("MultipleCallsConsistency", func(t *testing.T) {
		// Test multiple calls to ensure consistent caching behavior
		endpoint := "/health"
		numCalls := 10
		cacheHits := 0

		// First call to prime the cache
		dispatcher.Call("GET", endpoint, nil, nil, mockReq)

		for i := 0; i < numCalls; i++ {
			resp, err := dispatcher.Call("GET", endpoint, nil, nil, mockReq)
			if err != nil {
				t.Fatalf("Call %d failed: %v", i+1, err)
			}

			if resp.CacheHit {
				cacheHits++
			}
		}

		expectedHits := numCalls // All calls after the first should be cache hits
		if cacheHits != expectedHits {
			t.Errorf("Expected %d cache hits, got %d", expectedHits, cacheHits)
		}

		t.Logf("✅ Cache consistency: %d/%d cache hits", cacheHits, numCalls)
	})
}

// TestHTTPMethodHandling tests caching behavior for different HTTP methods
func TestHTTPMethodHandling(t *testing.T) {
	dispatcher, mockReq := setupTestEnvironment(t)

	methods := []struct {
		method      string
		shouldCache bool
	}{
		{"GET", true},
		{"POST", false},
		{"PUT", false},
		{"DELETE", false},
	}

	for _, test := range methods {
		t.Run(test.method, func(t *testing.T) {
			// First call
			_, err := dispatcher.Call(test.method, "/health", nil, nil, mockReq)
			if err != nil {
				t.Fatalf("First %s call failed: %v", test.method, err)
			}

			// Second call
			resp2, err := dispatcher.Call(test.method, "/health", nil, nil, mockReq)
			if err != nil {
				t.Fatalf("Second %s call failed: %v", test.method, err)
			}

			if test.shouldCache {
				if !resp2.CacheHit {
					t.Errorf("%s requests should be cached, but second call was not a cache hit", test.method)
				}
			} else {
				if resp2.CacheHit {
					t.Errorf("%s requests should not be cached, but second call was a cache hit", test.method)
				}
			}

			t.Logf("✅ %s caching behavior correct: cache hit = %v", test.method, resp2.CacheHit)
		})
	}
}

// TestConcurrentAccess tests thread safety of the caching system
func TestConcurrentAccess(t *testing.T) {
	dispatcher, mockReq := setupTestEnvironment(t)

	t.Run("ConcurrentCacheAccess", func(t *testing.T) {
		// Pre-warm the cache
		dispatcher.Call("GET", "/health", nil, nil, mockReq)

		numWorkers := 10
		results := make(chan bool, numWorkers)

		// Launch concurrent workers
		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				resp, err := dispatcher.Call("GET", "/health", nil, nil, mockReq)
				results <- (err == nil && resp.CacheHit)
			}(i)
		}

		// Collect results
		cacheHits := 0
		for i := 0; i < numWorkers; i++ {
			if <-results {
				cacheHits++
			}
		}

		if cacheHits != numWorkers {
			t.Errorf("Expected %d cache hits, got %d", numWorkers, cacheHits)
		}

		t.Logf("✅ Concurrent access: %d/%d cache hits (100%%)", cacheHits, numWorkers)
	})
}

// TestCacheStatistics tests cache statistics tracking
func TestCacheStatistics(t *testing.T) {
	dispatcher, mockReq := setupTestEnvironment(t)

	t.Run("StatisticsTracking", func(t *testing.T) {
		// Make some API calls
		dispatcher.Call("GET", "/health", nil, nil, mockReq)      // miss
		dispatcher.Call("GET", "/health", nil, nil, mockReq)      // hit
		dispatcher.Call("GET", "/admin", nil, nil, mockReq)       // miss
		dispatcher.Call("GET", "/admin", nil, nil, mockReq)       // hit
		dispatcher.Call("GET", "/nonexistent", nil, nil, mockReq) // miss (404)

		stats := dispatcher.GetCacheStats()

		if !stats["enabled"].(bool) {
			t.Error("Cache should be enabled")
		}

		totalEntries, ok := stats["total_entries"].(int)
		if !ok || totalEntries < 2 {
			t.Errorf("Expected at least 2 cache entries, got %v", totalEntries)
		}

		cacheHits, ok := stats["cache_hits"].(int)
		if !ok || cacheHits < 2 {
			t.Errorf("Expected at least 2 cache hits, got %v", cacheHits)
		}

		cacheMisses, ok := stats["cache_misses"].(int)
		if !ok || cacheMisses < 3 {
			t.Errorf("Expected at least 3 cache misses, got %v", cacheMisses)
		}

		t.Logf("✅ Cache statistics: entries=%d, hits=%d, misses=%d",
			totalEntries, cacheHits, cacheMisses)
	})
}

// BenchmarkCachedVsUncached benchmarks cached vs uncached performance
func BenchmarkCachedVsUncached(b *testing.B) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)
	mockReq.Header.Set("User-Agent", "Benchmark-Test")

	b.Run("Uncached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Use unique paths to avoid caching
			path := fmt.Sprintf("/health?uncached=%d", i)
			dispatcher.Call("GET", path, nil, nil, mockReq)
		}
	})

	b.Run("Cached", func(b *testing.B) {
		// Pre-warm cache
		dispatcher.Call("GET", "/health", nil, nil, mockReq)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dispatcher.Call("GET", "/health", nil, nil, mockReq)
		}
	})
}

// BenchmarkConcurrentAccess benchmarks concurrent cache access
func BenchmarkConcurrentAccess(b *testing.B) {
	sitesPath := common.GetEnv("SITES_PATH", "../sites")
	siteManager := common.NewSiteManager(sitesPath)
	renderEngine := common.NewRenderEngine(siteManager)
	r := routes.SetupRoutes(siteManager, renderEngine, false, 10, 100)
	dispatcher := common.NewRouterAPIDispatcher(r)

	mockReq, _ := http.NewRequest("GET", "http://localhost:8080/health", nil)
	mockReq.Header.Set("User-Agent", "Concurrent-Benchmark")

	// Pre-warm cache
	dispatcher.Call("GET", "/health", nil, nil, mockReq)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dispatcher.Call("GET", "/health", nil, nil, mockReq)
		}
	})
}
