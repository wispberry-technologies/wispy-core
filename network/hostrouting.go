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
		notFound = http.NotFoundHandler()
	}

	return &HostRouter{
		siteManager: siteManager,
		notFound:    notFound,
		defaultHost: defaultHost,
	}
}

// ServeHTTP implements http.Handler
func (hr *HostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract host from request
	host := r.Host

	// Remove port if present
	if idx := strings.IndexByte(host, ':'); idx >= 0 {
		host = host[:idx]
	}

	// Check for debug mode
	includeDebugInfo := false
	if r.URL.Query().Get("__include_debug_info__") == "true" || r.Header.Get("__include_debug_info__") == "true" {
		includeDebugInfo = true
	}

	// Try to get the site for this host
	s, err := hr.siteManager.GetSite(host)
	if err != nil {
		// If we have a default host, try that
		if hr.defaultHost != "" && host != hr.defaultHost {
			s, err = hr.siteManager.GetSite(hr.defaultHost)
		}

		// If we still don't have a site, use the not found handler
		if err != nil {
			common.Warning("No site found for host: %s", host)

			if includeDebugInfo {
				// Include debug info in the response
				http.Error(w, fmt.Sprintf("Site not found for host: %s\nError: %v", host, err), http.StatusNotFound)
				return
			}

			hr.notFound.ServeHTTP(w, r)
			return
		}
	}

	// Get the router for this site
	router := s.GetRouter()

	// If the router is nil, use the not found handler
	if router == nil {
		common.Warning("No router found for site: %s", host)

		if includeDebugInfo {
			// Include debug info in the response
			http.Error(w, fmt.Sprintf("No router found for site: %s", host), http.StatusInternalServerError)
			return
		}

		hr.notFound.ServeHTTP(w, r)
		return
	}

	// Serve the request using the site's router
	router.ServeHTTP(w, r)
}
