package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"

	"wispy-core/common"
)

// CreateDomainListFromMap creates a domain allowlist from a map and additional domains
func CreateDomainListFromMap(domainMap map[string]string, additionalDomains []string) []string {
	// Create domain allowlist for autocert
	domains := make([]string, 0, len(domainMap)+len(additionalDomains))

	// Add domain map entries
	for domain := range domainMap {
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	// Add additional domains
	for _, domain := range additionalDomains {
		if domain != "" && domain != "*" {
			domains = append(domains, domain)
		}
	}

	return domains
}

// NewSSLServer creates a new HTTPS server with autocert support
func NewSSLServer(certsDir, addr string, domains []string, handler http.Handler) (*autocert.Manager, *http.Server) {
	// Create autocert manager
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(certsDir),
		HostPolicy: hostPolicy(domains),
	}

	// Create HTTPS server
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			NextProtos: []string{
				"h2", "http/1.1",
				acme.ALPNProto, // Support TLS-ALPN-01 challenge
			},
			MinVersion: tls.VersionTLS12,
		},
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Return both manager and server
	return certManager, server
}

// hostPolicy returns a function that validates if a host is in the allowlist
func hostPolicy(domains []string) func(ctx context.Context, host string) error {
	hasWildcard := false
	for _, domain := range domains {
		if domain == "*" {
			hasWildcard = true
			break
		}
	}

	return func(ctx context.Context, host string) error {
		// If wildcard is present, accept all domains
		if hasWildcard {
			return nil
		}

		// Normalize host by removing port if present
		if idx := strings.IndexByte(host, ':'); idx >= 0 {
			host = host[:idx]
		}

		// Check if host is in allowlist
		for _, domain := range domains {
			if domain == host {
				return nil
			}
		}

		return fmt.Errorf("host %q not allowed", host)
	}
}

// StartACMEChallengeServer starts an HTTP server for ACME HTTP-01 challenges
func StartACMEChallengeServer(certManager *autocert.Manager, addr string) {
	if certManager == nil {
		return
	}

	// Create a server to handle ACME challenges
	httpServer := &http.Server{
		Addr:    addr,
		Handler: certManager.HTTPHandler(nil),
	}

	// Start HTTP server for ACME challenges
	go func() {
		common.Info("Starting ACME challenge server on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Warning("ACME challenge server failed: %v", err)
		}
	}()
}
