package network

import (
	"fmt"
	"net/http"
	"strings"

	"wispy-core/common"
	"wispy-core/core/site"
)

// HostRouter is a router that routes requests to the appropriate host based on the domain.
type HostRouter struct {
	siteManager site.SiteManager
	notFound    http.Handler
	defaultHost string
}

// NewHostRouter creates a new host router
func NewHostRouter(siteManager site.SiteManager, notFound http.Handler, defaultHost string) *HostRouter {
	// Validate inputs
	if siteManager == nil {
		panic("siteManager cannot be nil")
	}

	if notFound == nil {
		common.Warning("No notFound handler provided, using default http.NotFoundHandler")
		common.Warning("from: network.NewHostRouter()")
		notFound = http.NotFoundHandler()
	}

	return &HostRouter{
		siteManager: siteManager,
		notFound:    notFound,
		defaultHost: defaultHost,
	}
}

// extractHost removes any port number from the host string
func extractHost(hostWithPort string) string {
	if idx := strings.IndexByte(hostWithPort, ':'); idx >= 0 {
		return hostWithPort[:idx]
	}
	return hostWithPort
}

// extractDomainFromLocalhost extracts the actual domain from localhost subdomain patterns
// Example: "joosyjools.com.localhost" -> "joosyjools.com"
func extractDomainFromLocalhost(host string) string {
	if strings.HasSuffix(host, ".localhost") {
		return strings.TrimSuffix(host, ".localhost")
	}
	return host
}

// isDebugRequested checks if debug info was requested via query param or header
func isDebugRequested(r *http.Request) bool {
	return r.URL.Query().Get("__include_debug_info__") == "true" ||
		r.Header.Get("__include_debug_info__") == "true"
}

// respondWithNotFound responds with a not found error
func respondWithNotFound(w http.ResponseWriter, r *http.Request, host string, err error, includeDebug bool, handler http.Handler) {
	if includeDebug {
		http.Error(w, fmt.Sprintf("Site not found for host: %s\nError: %v", host, err), http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r)
}

// ServeHTTP implements http.Handler
func (hr *HostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := extractHost(r.Host)
	originalHost := host
	includeDebug := isDebugRequested(r)

	// Handle local development case
	if common.IsLocalDevelopment() {
		if strings.HasSuffix(host, ".localhost") {
			// Handle subdomain pattern like "joosyjools.com.localhost"
			extractedDomain := extractDomainFromLocalhost(host)
			common.Debug("Debug targeting host via subdomain: %s (from %s)", extractedDomain, originalHost)
			host = extractedDomain
		}
	}

	// Try to get the site for this host
	site, err := hr.siteManager.GetSite(host)

	// If site not found, try the default host (aka fallback to "localhost")
	if err != nil && hr.defaultHost != "" && host != hr.defaultHost {
		site, err = hr.siteManager.GetSite(hr.defaultHost)
	}

	// Still no site found, return not found
	if err != nil {
		common.Debug("No site found for host: %s, error: %v", host, err)
		common.Debug("defaultHost is set to: %s", hr.defaultHost)
		common.Debug("domains: %v", hr.siteManager.Domains().GetDomains())
		respondWithNotFound(w, r, host, err, includeDebug, hr.notFound)
		return
	}

	// Get the router for this site
	router := site.GetRouter()

	// If the router is nil, return not found with a different error
	if router == nil {
		routerErr := fmt.Errorf("no router configured for site")
		common.Error(routerErr.Error())
		// Return request too soon, no router configured for site
		// This is a common case when the site is not fully initialized or the router is not set up yet.
		http.Error(w, "Site is not ready", http.StatusTooEarly)
		return
	}

	// Serve the request using the site's router
	router.ServeHTTP(w, r)
}
