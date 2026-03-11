package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/shivamk01here/distributed-gateway/internal/api"
	"github.com/shivamk01here/distributed-gateway/internal/limiter"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-key-for-local-testing"
	}

	log.Printf("Connecting to Redis at %s...", redisAddr)
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	log.Println("Initializing Sliding Window Rate Limiter...")
	slidingLimiter := limiter.NewSlidingWindowLimiter(rdb, 5, 10)

	log.Println("Configuring Load Balancer...")
	backendURLs := []string{"http://backend_alpha:8080", "http://backend_beta:8080"}
	lb := api.NewLoadBalancer(backendURLs)

	finalHandler := lb.ReverseProxyHandler()

	rateLimitedHandler := api.RateLimitMiddleware(slidingLimiter, finalHandler)

	secureHandler := api.AuthMiddleware(jwtSecret, rateLimitedHandler)

	http.HandleFunc("/api/", secureHandler)

	http.Handle("/metrics", promhttp.Handler())

	port := ":8080"
	log.Printf("Distributed API Gateway live on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Gateway crashed: %v", err)
	}
}
