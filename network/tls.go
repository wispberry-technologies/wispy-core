package network

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
	"wispy-core/common"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// SSLValidator interface for managing domain certificates in memory and on disk.
type SSLValidator interface {
	AddDomain(domain string) error
	HasCert(domain string) bool
	Info(domain string) (exists bool, cached bool)
	GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)
}

// sslValidatorImpl implements SSLValidator.
type sslValidatorImpl struct {
	certsDir  string
	domains   DomainList
	certCache map[string]*tls.Certificate
	mu        sync.RWMutex
}

// NewSSLValidator creates a new SSLValidator with the given certsDir and allowed domains.
func NewSSLValidator(certsDir string, domains DomainList) SSLValidator {
	return &sslValidatorImpl{
		certsDir:  certsDir,
		domains:   domains,
		certCache: make(map[string]*tls.Certificate),
	}
}

func (s *sslValidatorImpl) AddDomain(domain string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.domains.AddDomain(domain)
	return nil
}

func (s *sslValidatorImpl) HasCert(domain string) bool {
	s.mu.RLock()
	_, ok := s.certCache[domain]
	s.mu.RUnlock()
	if ok {
		return true
	}
	// Try loading from disk if not in cache
	certPath := filepath.Join(s.certsDir, domain+".crt")
	keyPath := filepath.Join(s.certsDir, domain+".key")
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err == nil {
		s.mu.Lock()
		s.certCache[domain] = &cert
		s.mu.Unlock()
		return true
	}
	return false
}

func (s *sslValidatorImpl) Info(domain string) (exists bool, cached bool) {
	s.mu.RLock()
	exists = s.domains.HasDomain(domain)
	cached = s.certCache[domain] != nil
	s.mu.RUnlock()
	return exists, cached
}

func (s *sslValidatorImpl) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := hello.ServerName
	s.mu.RLock()
	allowed := s.domains.HasDomain(domain)
	s.mu.RUnlock()
	if !allowed {
		return nil, fmt.Errorf("domain %s not allowed", domain)
	}

	// Check cache
	s.mu.RLock()
	cert, ok := s.certCache[domain]
	s.mu.RUnlock()
	if ok {
		return cert, nil
	}

	// Try loading from disk
	certPath := filepath.Join(s.certsDir, domain+".crt")
	keyPath := filepath.Join(s.certsDir, domain+".key")
	loadedCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err == nil {
		s.mu.Lock()
		s.certCache[domain] = &loadedCert
		s.mu.Unlock()
		return &loadedCert, nil
	}

	// If localhost or IP, generate self-signed cert
	if isLocalhostOrIP(domain) {
		selfCert, err := generateSelfSignedCert(domain, s.certsDir)
		if err != nil {
			return nil, fmt.Errorf("failed to generate self-signed cert for %s: %w", domain, err)
		}
		s.mu.Lock()
		s.certCache[domain] = selfCert
		s.mu.Unlock()
		return selfCert, nil
	}

	// Generate and save cert via Let's Encrypt if not present and valid DNS name
	if err := generateAndSaveCert(domain, s.certsDir); err != nil {
		return nil, fmt.Errorf("failed to generate cert for %s: %w", domain, err)
	}
	loadedCert, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load cert after generation for %s: %w", domain, err)
	}
	s.mu.Lock()
	s.certCache[domain] = &loadedCert
	s.mu.Unlock()
	return &loadedCert, nil
}

// isLocalhostOrIP returns true if the domain is "localhost" or an IP address.
func isLocalhostOrIP(domain string) bool {
	if domain == "localhost" {
		return true
	}
	ip := net.ParseIP(domain)
	return ip != nil
}

// generateSelfSignedCert generates a self-signed certificate for localhost or an IP and saves it to disk.
func generateSelfSignedCert(domain, certsDir string) (*tls.Certificate, error) {
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create certs dir: %w", err)
	}
	certPath := filepath.Join(certsDir, domain+".crt")
	keyPath := filepath.Join(certsDir, domain+".key")

	// If already exists, load and return
	if _, err := os.Stat(certPath); err == nil {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err == nil {
			return &cert, nil
		}
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: domain},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if domain == "localhost" {
		template.DNSNames = []string{"localhost"}
		template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	} else if ip := net.ParseIP(domain); ip != nil {
		template.IPAddresses = []net.IP{ip}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certOut, err := os.Create(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cert file for writing: %w", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, fmt.Errorf("failed to write cert: %w", err)
	}

	keyOut, err := os.Create(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file for writing: %w", err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal private key: %w", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, fmt.Errorf("failed to write key: %w", err)
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load generated self-signed cert: %w", err)
	}
	return &cert, nil
}

// generateAndSaveCert obtains a production certificate for the given domain
// using Let's Encrypt and saves the cert and key to certsDir/{domain}.crt and .key.
func generateAndSaveCert(domain, certsDir string) error {
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return fmt.Errorf("failed to create certs dir: %w", err)
	}

	manager := autocert.Manager{
		Cache:      autocert.DirCache(certsDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Email:      "",
	}

	// Use staging directory if requested
	if common.IsProduction() {
		manager.Client = &acme.Client{
			DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
		}
	} else {
		manager.Client = &acme.Client{
			DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		}
	}

	tlsCert, err := manager.GetCertificate(&tls.ClientHelloInfo{ServerName: domain})
	if err != nil {
		return fmt.Errorf("failed to obtain certificate from Let's Encrypt: %w", err)
	}

	certOut, err := os.Create(filepath.Join(certsDir, domain+".crt"))
	if err != nil {
		return fmt.Errorf("failed to open cert file for writing: %w", err)
	}
	defer certOut.Close()
	for _, certDER := range tlsCert.Certificate {
		if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
			return fmt.Errorf("failed to write cert: %w", err)
		}
	}

	keyOut, err := os.Create(filepath.Join(certsDir, domain+".key"))
	if err != nil {
		return fmt.Errorf("failed to open key file for writing: %w", err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(tlsCert.PrivateKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %w", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write key: %w", err)
	}

	return nil
}
