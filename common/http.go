package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GetIPAddress(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return strings.Split(xff, ",")[0]
	}

	// Fallback to RemoteAddr if no X-Forwarded-For header is present
	ip := r.RemoteAddr
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx] // Strip port if present
	}
	return ip
}

// shouldIncludeDebugInfo checks if debug info should be included in the response
// based on project standards:
// - Query parameter: __include_debug_info__=true
// - HTTP header: __include_debug_info__: true
func ShouldIncludeDebugInfo(r *http.Request) bool {
	// Check query parameter
	if r.URL.Query().Get("__include_debug_info__") == "true" {
		return true
	}

	// Check header
	return r.Header.Get("__include_debug_info__") == "true"
}

// respondWithError writes a plain text error response based on project standards
// All API error responses MUST be in plain text
func RespondWithError(w http.ResponseWriter, r *http.Request, status int, message string, err error) {
	// Generate debug info if requested
	var debugInfo string
	if err != nil && ShouldIncludeDebugInfo(r) {
		debugInfo = fmt.Sprintf("%v", err)
	}

	// Use the common PlainTextError function
	PlainTextError(w, status, message, debugInfo)

	// Log the error
	if err != nil {
		Error("Auth error: %s - %v", message, err)
	} else {
		Info("Auth info: %s", message)
	}
}

// RespondWithJSON writes a JSON response with the given status code
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

// RedirectWithMessage redirects to a URL with message and optionally error parameters
func RedirectWithMessage(w http.ResponseWriter, r *http.Request, redirectURL string, message string, errMsg string) {
	if message != "" {
		if redirectURL == "" {
			redirectURL = "/"
		}

		if strings.Contains(redirectURL, "?") {
			redirectURL += "&message=" + url.QueryEscape(message)
		} else {
			redirectURL += "?message=" + url.QueryEscape(message)
		}

		if errMsg != "" {
			redirectURL += "&error=" + url.QueryEscape(errMsg)
		}
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// PlainTextError writes a plain text error response with status code and X-Debug header
func PlainTextError(w http.ResponseWriter, status int, msg string, debug ...string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if len(debug) > 0 {
		w.Header().Add("X-Debug", debug[0])
	}
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func RespondWithPlainText(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(msg)); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", err), http.StatusInternalServerError)
		return
	}
}

func NormalizeHost(host string) string {
	// If the host contains a port, strip it
	h := strings.Split(host, ":")[0]

	// If the host is localhost, return the default host
	if h == "localhost" || h == "127.0.0.1" {
		return "localhost"
	}

	return h
}

func Redirect404(w http.ResponseWriter, r *http.Request, url string) {
	// If the URL is empty, redirect to the root
	if url == "" {
		url = "/404?original_url=" + r.URL.Path
	}

	// Set the Location header and status code
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)

	// Log the redirect
	Debug("Redirecting %s to %s", r.URL.Path, url)
}

func RequestLogger() func(http.Handler) http.Handler {
	// Example output:
	// Jun 21 11:02:05.195 "GET http://localhost:8080/login HTTP/1.1" from 127.0.0.1:60741 - 200 816530B in 17.427083ms
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			ServerLog("Request: %s %s from %s - %d %s in %v",
				r.Method, r.URL.Path, r.RemoteAddr,
				http.StatusOK, http.StatusText(http.StatusOK), duration)
		})
	}
}
