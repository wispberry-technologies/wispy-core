package core

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"wispy-core/pkg/common"
)

// NewReverseProxyHandler returns an http.Handler that proxies requests to the given targetURL
func NewReverseProxyHandler(targetURL string, routePrefix string) http.HandlerFunc {
	// Add scheme if missing (default to http)

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		common.Fatal("Failed to parse target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	// Optional: error callback for logging
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		// You can replace this with your logger
		http.Error(w, "Proxy error: "+e.Error(), http.StatusBadGateway)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		common.Debug("Proxying request: %s %s to %s", r.Method, r.URL.Path, targetURL)
		// Rewrite the request URL path to strip the routePrefix
		if strings.HasPrefix(r.URL.Path, routePrefix) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, routePrefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		r.Header.Set("X-Wispy-Api-Key", common.GetEnv("WISPY_CORE_PROXY_KEY", "no-key-found"))
		proxy.ServeHTTP(w, r)
	}
}
