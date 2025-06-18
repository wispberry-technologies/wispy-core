package common

import (
	"net/http"
	"strings"
)

// PlainTextError writes a plain text error response with status code and X-Debug header
func PlainTextError(w http.ResponseWriter, status int, msg string, debug ...string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if len(debug) > 0 {
		w.Header().Add("X-Debug", debug[0])
	}
	w.WriteHeader(status)
	w.Write([]byte(msg))
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
