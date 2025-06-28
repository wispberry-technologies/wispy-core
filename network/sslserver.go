package network

import (
	"crypto/tls"
	"net/http"
	"time"
	"wispy-core/common"
)

func NewSSLServer(certsDir string, serverAddress string, domains DomainList, router http.Handler) (sslManager SSLValidator, server *http.Server) {
	sslManager = NewSSLValidator(certsDir, domains)

	isProduction := common.IsProduction()
	isDevelopment := !isProduction

	server = &http.Server{
		Addr: serverAddress,
		TLSConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			GetCertificate:     sslManager.GetCertificate,
			InsecureSkipVerify: isDevelopment, // Allow insecure connections in development mode
		},
		Handler:        router,
		IdleTimeout:    30 * time.Second,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return sslManager, server
}
