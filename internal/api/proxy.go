package api

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewReverseProxy creates a handler that forwards requests to the target backend.
func NewReverseProxy(targetHost string) http.HandlerFunc {
	url, err := url.Parse(targetHost)
	if err != nil {
		log.Fatalf("Invalid backend URL: %v", err)
	}

	// Go's built-in reverse proxy handles (headers, chunking, etc.)
	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host

		log.Printf("[PROXY] Forwarding request to %s%s", targetHost, r.URL.Path)
		proxy.ServeHTTP(w, r)
	}
}
