package network

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

// NewLocalSSLServer creates a new HTTPS server for local development with self-signed certificates
func NewLocalSSLServer(certsDir, addr string, handler http.Handler) (*http.Server, error) {
	// Ensure local certificates exist
	cert, err := ensureLocalCerts(certsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure local certificates: %w", err)
	}

	// Create TLS config with the local certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2", "http/1.1"},
	}

	// Create HTTPS server
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server, nil
}

// ensureLocalCerts ensures that local development certificates exist and returns them
func ensureLocalCerts(certsDir string) (*tls.Certificate, error) {
	certPath := filepath.Join(certsDir, "local.crt")
	keyPath := filepath.Join(certsDir, "local.key")

	// Check if certificates already exist and are valid
	if cert, err := loadAndValidateLocalCert(certPath, keyPath); err == nil {
		return cert, nil
	}

	// Generate new local certificates
	if err := generateLocalCerts(certsDir); err != nil {
		return nil, fmt.Errorf("failed to generate local certificates: %w", err)
	}

	// Load the newly generated certificates
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load local certificates: %w", err)
	}

	common.Info("Generated new local development certificates")
	return &cert, nil
}

// loadAndValidateLocalCert loads and validates existing local certificates
func loadAndValidateLocalCert(certPath, keyPath string) (*tls.Certificate, error) {
	// Check if files exist
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("certificate file does not exist")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("key file does not exist")
	}

	// Load the certificate
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	// Parse the certificate to check expiration
	if len(cert.Certificate) > 0 {
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %w", err)
		}

		// Check if certificate is expired or expires within 30 days
		if time.Now().After(x509Cert.NotAfter) || time.Now().Add(30*24*time.Hour).After(x509Cert.NotAfter) {
			return nil, fmt.Errorf("certificate is expired or expires soon")
		}
	}

	return &cert, nil
}

// generateLocalCerts generates self-signed certificates for local development
func generateLocalCerts(certsDir string) error {
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	certPath := filepath.Join(certsDir, "local.crt")
	keyPath := filepath.Join(certsDir, "local.key")

	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"Wispy Core Local Development"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Local"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames: []string{
			"localhost",
			"*.localhost",
			"local.dev",
			"*.local.dev",
			"*.test",
			"*.local",
		},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
			net.ParseIP("0.0.0.0"),
		},
	}

	// Create the certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate file
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write key file
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}
