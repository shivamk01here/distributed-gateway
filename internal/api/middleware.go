package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

func RateLimitMiddleware(rl limiter.RateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			ip = strings.Split(forwardedFor, ",")[0]
		}

		if !rl.Allow(ip) {
			log.Printf("[THROTTLED] IP: %s", ip)
			GatewayRequestsTotal.WithLabelValues("rate_limited").Inc()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		GatewayRequestsTotal.WithLabelValues("allowed").Inc()
		next.ServeHTTP(w, r)
	}
}
