// File: internal/api/middleware.go
package api

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

// RateLimitMiddleware for intercepts incoming requests and applies the rate limiter.
func RateLimitMiddleware(rl limiter.RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the IP address of the client to use as the rate-limiting key.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("Error extracting IP: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			ip = strings.Split(forwardedFor, ",")[0]
		}

		if !rl.Allow(ip) {
			log.Printf("[THROTTLED] Request rejected for IP: %s", ip)
			// Return a 429 Too Many Requests status code
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		// If allowed, log the success and pass the request to the actual handler
		log.Printf("[ACCEPTED] Request allowed for IP: %s", ip)
		next.ServeHTTP(w, r)
	}
}
