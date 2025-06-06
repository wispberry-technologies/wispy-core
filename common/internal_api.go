package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"
)

// APIDispatcher interface for internal API calls
type APIDispatcher interface {
	Call(method, path string, body []byte, headers map[string]string, originalRequest *http.Request) (*APIResponse, error)
}

// APIResponse represents the response from an internal API call
type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       string                 `json:"body"`
	Data       map[string]interface{} `json:"data,omitempty"`      // Parsed JSON data if applicable
	Duration   time.Duration          `json:"duration,omitempty"`  // Call duration for performance monitoring
	CacheHit   bool                   `json:"cache_hit,omitempty"` // Whether this was served from cache
}

// CachedResponse represents a cached API response
type CachedResponse struct {
	Response  *APIResponse
	Timestamp time.Time
	TTL       time.Duration
}

// APICache provides simple in-memory caching for API responses
type APICache struct {
	cache  map[string]*CachedResponse
	mutex  sync.RWMutex
	hits   int
	misses int
}

// NewAPICache creates a new API cache
func NewAPICache() *APICache {
	return &APICache{
		cache:  make(map[string]*CachedResponse),
		mutex:  sync.RWMutex{},
		hits:   0,
		misses: 0,
	}
}

// Get retrieves a cached response if it exists and is not expired
func (c *APICache) Get(key string) (*APIResponse, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	cached, exists := c.cache[key]
	if !exists {
		c.misses++
		return nil, false
	}

	// Check if expired
	if time.Since(cached.Timestamp) > cached.TTL {
		c.misses++
		return nil, false
	}

	c.hits++
	// Mark as cache hit
	response := *cached.Response
	response.CacheHit = true
	return &response, true
}

// Set stores a response in cache with TTL
func (c *APICache) Set(key string, response *APIResponse, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &CachedResponse{
		Response:  response,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// Clear removes expired entries from cache
func (c *APICache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, cached := range c.cache {
		if now.Sub(cached.Timestamp) > cached.TTL {
			delete(c.cache, key)
		}
	}
}

// Size returns the number of cached responses
func (c *APICache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache)
}

// GetStats returns cache statistics
func (c *APICache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"total_entries": len(c.cache),
		"cache_hits":    c.hits,
		"cache_misses":  c.misses,
	}
}

// RouterAPIDispatcher implements APIDispatcher using a chi router
type RouterAPIDispatcher struct {
	router http.Handler
	cache  *APICache
}

// NewRouterAPIDispatcher creates a new RouterAPIDispatcher
func NewRouterAPIDispatcher(router http.Handler) *RouterAPIDispatcher {
	return &RouterAPIDispatcher{
		router: router,
		cache:  NewAPICache(),
	}
}

// NewRouterAPIDispatcherWithCache creates a new RouterAPIDispatcher with custom cache
func NewRouterAPIDispatcherWithCache(router http.Handler, cache *APICache) *RouterAPIDispatcher {
	return &RouterAPIDispatcher{
		router: router,
		cache:  cache,
	}
}

// Call executes an internal API call through the router
func (d *RouterAPIDispatcher) Call(method, path string, body []byte, headers map[string]string, originalRequest *http.Request) (*APIResponse, error) {
	startTime := time.Now()

	// Validate inputs
	if method == "" {
		return nil, fmt.Errorf("HTTP method cannot be empty")
	}
	if path == "" {
		return nil, fmt.Errorf("API path cannot be empty")
	}
	if d.router == nil {
		return nil, fmt.Errorf("router is not initialized")
	}

	// Normalize method to uppercase
	method = strings.ToUpper(method)

	// For GET requests, check cache first
	var cacheKey string
	if method == "GET" && d.cache != nil {
		cacheKey = generateCacheKey(method, path, headers)
		if cached, hit := d.cache.Get(cacheKey); hit {
			cached.Duration = time.Since(startTime)
			return cached, nil
		}
	}

	// Create a new request
	req, err := http.NewRequest(method, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s %s: %w", method, path, err)
	}

	// Copy relevant headers from original request
	if originalRequest != nil {
		// Forward authentication and session headers
		forwardHeaders := []string{
			"Authorization",
			"Cookie",
			"X-Forwarded-For",
			"X-Real-IP",
			"User-Agent",
			"X-Session-ID",
			"X-Site-Domain", // Custom header for multi-site context
			"Accept",
			"Accept-Language",
			"Accept-Encoding",
		}

		for _, headerName := range forwardHeaders {
			if value := originalRequest.Header.Get(headerName); value != "" {
				req.Header.Set(headerName, value)
			}
		}

		// Copy context from original request (preserves authentication, tracing, etc.)
		req = req.WithContext(originalRequest.Context())

		// Copy remote address for IP-based security
		req.RemoteAddr = originalRequest.RemoteAddr

		// Copy host information for multi-tenant routing
		req.Host = originalRequest.Host

		// Copy URL query parameters if not already in path
		if originalRequest.URL != nil && originalRequest.URL.RawQuery != "" && !strings.Contains(path, "?") {
			// Only copy if the internal path doesn't already have query params
			if parsedURL, err := url.Parse(path); err == nil && parsedURL.RawQuery == "" {
				req.URL.RawQuery = originalRequest.URL.RawQuery
			}
		}
	}

	// Add custom headers (these override forwarded headers if there's a conflict)
	for key, value := range headers {
		if key != "" && value != "" {
			req.Header.Set(key, value)
		}
	}

	// Set content type for body if not already set
	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add internal API marker to distinguish from external requests
	req.Header.Set("X-Internal-API", "true")
	req.Header.Set("X-Internal-API-Timestamp", startTime.Format(time.RFC3339))

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Execute the request through the router
	d.router.ServeHTTP(recorder, req)

	// Calculate duration
	duration := time.Since(startTime)

	// Parse response headers
	responseHeaders := make(map[string]string)
	for key, values := range recorder.Header() {
		if len(values) > 0 {
			responseHeaders[key] = strings.Join(values, ", ")
		}
	}

	// Create API response
	apiResponse := &APIResponse{
		StatusCode: recorder.Code,
		Headers:    responseHeaders,
		Body:       recorder.Body.String(),
		Duration:   duration,
		CacheHit:   false,
	}

	// Try to parse JSON response with better error handling
	contentType := strings.ToLower(responseHeaders["Content-Type"])
	if strings.Contains(contentType, "application/json") {
		var data map[string]interface{}
		bodyBytes := recorder.Body.Bytes()

		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &data); err == nil {
				apiResponse.Data = data
			} else {
				// If JSON parsing fails, still provide the raw body
				// This helps with debugging malformed JSON responses
				apiResponse.Data = map[string]interface{}{
					"_parse_error": fmt.Sprintf("Failed to parse JSON: %v", err),
					"_raw_body":    string(bodyBytes),
				}
			}
		} else {
			// Empty response body but JSON content type
			apiResponse.Data = map[string]interface{}{}
		}
	} else if strings.Contains(contentType, "text/") || contentType == "" {
		// For text responses, make the body easily accessible
		apiResponse.Data = map[string]interface{}{
			"text": apiResponse.Body,
		}
	}

	// Add response metadata for debugging
	apiResponse.Data = mergeMapSafely(apiResponse.Data, map[string]interface{}{
		"_internal_api": true,
		"_method":       method,
		"_path":         path,
		"_status_code":  recorder.Code,
		"_duration_ms":  duration.Milliseconds(),
		"_cache_hit":    false,
	})

	// Cache successful GET responses
	if method == "GET" && d.cache != nil && recorder.Code < 400 {
		cacheTTL := determineCacheTTL(responseHeaders)

		if cacheTTL > 0 {
			d.cache.Set(cacheKey, apiResponse, cacheTTL)
		}
	}

	return apiResponse, nil
}

// generateCacheKey creates a consistent cache key from request parameters
func generateCacheKey(method, path string, headers map[string]string) string {
	// Simple cache key generation - could be enhanced with hashing for large headers
	headersStr := ""
	if len(headers) > 0 {
		headerBytes, _ := json.Marshal(headers)
		headersStr = string(headerBytes)
	}
	return fmt.Sprintf("%s:%s:%s", method, path, headersStr)
}

// determineCacheTTL determines the cache TTL based on response headers
func determineCacheTTL(headers map[string]string) time.Duration {
	defaultTTL := 5 * time.Minute

	// Check for custom cache control headers
	if cacheControl := headers["Cache-Control"]; cacheControl != "" {
		// Simple cache control parsing (could be enhanced)
		if strings.Contains(cacheControl, "no-cache") || strings.Contains(cacheControl, "no-store") {
			return 0 // Don't cache
		}
		if strings.Contains(cacheControl, "max-age=") {
			// Extract max-age value (simple parsing)
			parts := strings.Split(cacheControl, "max-age=")
			if len(parts) > 1 {
				ageStr := strings.Split(parts[1], ",")[0]
				ageStr = strings.TrimSpace(ageStr)
				if duration, err := time.ParseDuration(ageStr + "s"); err == nil {
					return duration
				}
			}
		}
	}

	// Check for Expires header
	if expires := headers["Expires"]; expires != "" {
		if expTime, err := time.Parse(time.RFC1123, expires); err == nil {
			if ttl := time.Until(expTime); ttl > 0 {
				return ttl
			}
		}
	}

	return defaultTTL
}

// Helper function to safely merge maps
func mergeMapSafely(target map[string]interface{}, source map[string]interface{}) map[string]interface{} {
	if target == nil {
		target = make(map[string]interface{})
	}
	for k, v := range source {
		// Only add if key doesn't exist to avoid overwriting actual data
		if _, exists := target[k]; !exists {
			target[k] = v
		}
	}
	return target
}

// MockAPIDispatcher for testing
type MockAPIDispatcher struct {
	responses map[string]*APIResponse
	mutex     sync.RWMutex
}

// NewMockAPIDispatcher creates a new mock API dispatcher
func NewMockAPIDispatcher() *MockAPIDispatcher {
	return &MockAPIDispatcher{
		responses: make(map[string]*APIResponse),
		mutex:     sync.RWMutex{},
	}
}

// SetResponse sets a mock response for a given path
func (m *MockAPIDispatcher) SetResponse(method, path string, response *APIResponse) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	key := strings.ToUpper(method) + " " + path
	m.responses[key] = response
}

// Call returns a mock response
func (m *MockAPIDispatcher) Call(method, path string, body []byte, headers map[string]string, originalRequest *http.Request) (*APIResponse, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	key := strings.ToUpper(method) + " " + path
	if response, exists := m.responses[key]; exists {
		// Create a copy to avoid modifying the original
		responseCopy := *response
		responseCopy.Duration = time.Millisecond * 10 // Mock duration
		return &responseCopy, nil
	}

	return &APIResponse{
		StatusCode: 404,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"error": "Mock endpoint not found"}`,
		Data: map[string]interface{}{
			"error":        "Mock endpoint not found",
			"_method":      strings.ToUpper(method),
			"_path":        path,
			"_status_code": 404,
			"_duration_ms": int64(10),
		},
		Duration: time.Millisecond * 10,
	}, nil
}

// Helper function to parse URL and extract query parameters
func parseURLParams(urlStr string) (string, url.Values, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", nil, err
	}
	return u.Path, u.Query(), nil
}

// CleanupCache performs periodic cleanup of expired cache entries
func (d *RouterAPIDispatcher) CleanupCache() {
	if d.cache != nil {
		d.cache.Clear()
	}
}

// GetCacheStats returns cache statistics
func (d *RouterAPIDispatcher) GetCacheStats() map[string]interface{} {
	if d.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := d.cache.GetStats()
	stats["enabled"] = true
	return stats
}
